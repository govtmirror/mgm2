package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/googollee/go-socket.io"
	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/client"
	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/core/job"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/core/user"
	"github.com/m-o-s-e-s/mgm/email"
	"github.com/m-o-s-e-s/mgm/simian"
	"github.com/m-o-s-e-s/mgm/sql"
	"github.com/m-o-s-e-s/mgm/webClient"
	"github.com/satori/go.uuid"

	//"github.com/m-o-s-e-s/mgm/opensim"
	"flag"
	"net/http"

	"code.google.com/p/gcfg"
	"github.com/jcelliott/lumber"
)

func main() {
	//instantiate our logger
	logger := lumber.NewConsoleLogger(lumber.DEBUG)
	cfgPtr := flag.String("config", "/opt/mgm/mgm.gcfg", "path to config file")
	flag.Parse()

	logger.Info("Reading configuration file")
	//read configuration file
	config := core.MgmConfig{}
	err := gcfg.ReadFileInto(&config, *cfgPtr)
	if err != nil {
		logger.Fatal("Error reading config file: ", err)
		return
	}

	//instantiate our email module
	mailer := email.NewClientMailer(config.Email, config.Web.Hostname)

	logger.Info("Connecting to the database")
	//create our database connector
	db := sql.NewDatabase(
		config.MySQL.Username,
		config.MySQL.Password,
		config.MySQL.Database,
		config.MySQL.Host,
	)
	err = db.TestConnection()
	if err != nil {
		logger.Error("Connecting to mysql: ", err)
		return
	}
	osdb := sql.NewDatabase(
		config.Opensim.Username,
		config.Opensim.Password,
		config.Opensim.Database,
		config.Opensim.Host,
	)
	err = osdb.TestConnection()
	if err != nil {
		logger.Error("Connecting to opensim mysql: ", err)
		return
	}
	//create our simian connector
	sim, err := simian.NewConnector(config.MGM.SimianURL)
	if err != nil {
		logger.Error("Error instantiating Simian connection: ", err)
		return
	}

	//instantiate our persistance handler
	pers := sql.NewMGMDB(db, osdb, sim, logger)

	//perform any necessary migrations
	pers.Migrate(config.MGM.FilesDirectory)

	//create our client notifier
	notifier := client.NewNotifier()

	logger.Info("Populating caches")
	//Hook up core processing...
	jMgr := job.NewManager(config.Web.FileStorage, config.MGM.MgmURL, config.MGM.HubRegionUUID, pers, notifier, logger)
	rMgr := region.NewManager(config.MGM.MgmURL, config.MGM.SimianURL, pers, osdb, notifier, logger)
	hMgr := host.NewManager(config.MGM.NodePort, rMgr, pers, notifier, logger)
	uMgr := user.NewManager(rMgr, hMgr, jMgr, sim, pers, notifier, logger)

	cMgr := client.NewManager(uMgr, hMgr, rMgr, jMgr, notifier, logger)

	// http function handler
	httpCon := webClient.NewHTTPConnector(jMgr, pers, sim, uMgr, mailer, logger)

	//Create a socket.io websocket server to listed for client connections
	cServer, err := socketio.NewServer(nil)
	if err != nil {
		logger.Fatal("Error creating websocket server: ", err)
		return
	}
	//We have a connecting client
	cServer.On("connection", func(so socketio.Socket) {
		// a new client has connected

		log.Println("on connection")

		//query client for their jwt token
		so.Emit("Auth Challenge")
		so.On("Auth Response", func(inputToken string) {

			//validate token
			token, err := jwt.Parse(inputToken, func(token *jwt.Token) (interface{}, error) {
				if token.Method.Alg() == "HS256" {
					return []byte(config.MGM.SecretKey), nil
				}
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			})

			if err == nil && token.Valid {
				//Valid token, accept socket
				uid, _ := uuid.FromString(token.Claims["guid"].(string))
				cMgr.NewClient(so, uid)
			} else {
				//invalid token, deny socket
				logger.Info("token invalid ", err.Error())
				so.Emit("disconnect")
			}
		})
	})
	cServer.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	/***** Authentication Handler *****/
	loginHandler := func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		type clientAuthRequest struct {
			Username string
			Password string
		}
		type clientAuthResponse struct {
			UUID        uuid.UUID
			AccessLevel uint8
			Message     string
			Token       string
			Success     bool
		}

		var t clientAuthRequest
		err := decoder.Decode(&t)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err.Error()), http.StatusInternalServerError)
			return
		}

		u, valid := uMgr.Auth(t.Username, t.Password)
		if valid == false {
			http.Error(w, "Invalid Credential", http.StatusInternalServerError)
			return
		}

		//create an authentication token
		token := jwt.New(jwt.SigningMethodHS256)
		token.Claims["guid"] = u.UserID
		token.Claims["exp"] = time.Now().Add(time.Minute * 60).Unix()

		tokenString, err := token.SignedString([]byte(config.MGM.SecretKey))
		if err != nil {
			logger.Error("Error in Auth: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := clientAuthResponse{u.UserID, u.AccessLevel, "", tokenString, true}
		js, err := json.Marshal(response)
		if err != nil {
			logger.Error("Error in Auth: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

	mux := http.NewServeMux()
	mux.Handle("/ws/", cServer)
	mux.HandleFunc("/host/", hMgr.WShandler)
	mux.HandleFunc("/auth", loginHandler)
	mux.HandleFunc("/auth/register", cMgr.RegisterHandler)
	mux.HandleFunc("/auth/passwordToken", httpCon.PasswordTokenHandler)
	mux.HandleFunc("/auth/passwordReset", httpCon.PasswordResetHandler)
	mux.HandleFunc("/upload/{id}", httpCon.UploadHandler)
	mux.HandleFunc("/download/{id}", httpCon.DownloadHandler)
	mux.Handle("/", http.FileServer(http.Dir(config.Web.Root)))
	logger.Info("Listening for clients on :%d", config.MGM.WebPort)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.MGM.WebPort), mux); err != nil {
		logger.Fatal("ListenAndServe:", err)
	}
}
