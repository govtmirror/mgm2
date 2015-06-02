package main

import (
	"strconv"

	"github.com/gorilla/mux"
	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/core/job"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/core/session"
	"github.com/m-o-s-e-s/mgm/core/user"
	"github.com/m-o-s-e-s/mgm/email"
	"github.com/m-o-s-e-s/mgm/simian"
	"github.com/m-o-s-e-s/mgm/webClient"

	//"github.com/m-o-s-e-s/mgm/opensim"
	"flag"
	"net/http"

	"code.google.com/p/gcfg"
	"github.com/jcelliott/lumber"
)

type mgmConfig struct {
	MGM struct {
		SimianURL        string
		SessionSecret    string
		OpensimPort      string
		WebPort          int
		NodePort         int
		PublicHostname   string
		LocalFileStorage string
	}

	MySQL struct {
		Username string
		Password string
		Host     string
		Database string
	}

	Email email.EmailConfig
}

func main() {
	//instantiate our logger
	logger := logger.Wrap("MGM", lumber.NewConsoleLogger(lumber.DEBUG))

	cfgPtr := flag.String("config", "/opt/mgm/mgm.gcfg", "path to config file")

	flag.Parse()

	//read configuration file
	config := mgmConfig{}
	err := gcfg.ReadFileInto(&config, *cfgPtr)
	if err != nil {
		logger.Fatal("Error reading config file: ", err)
		return
	}

	//instantiate our email module
	mailer := email.NewClientMailer(config.Email, config.MGM.PublicHostname)

	//create our database connector
	db := database.NewDatabase(
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
	//create our simian connector
	sim, err := simian.NewConnector(config.MGM.SimianURL)
	if err != nil {
		logger.Error("Error instantiating Simian connection: ", err)
		return
	}

	//Hook up core processing...
	jMgr := job.NewManager(config.MGM.LocalFileStorage, db, logger)
	rMgr := region.NewManager(db, logger)
	nMgr, err := host.NewManager(config.MGM.NodePort, rMgr, db, logger)
	if err != nil {
		logger.Error("Error instantiating host manager: ", err)
		return
	}
	uMgr := user.NewManager(rMgr, nMgr, sim, db, logger)
	sessionListenerChan := make(chan core.UserSession, 64)

	_ = session.NewManager(sessionListenerChan, uMgr, jMgr, nMgr, rMgr, sim, logger)

	httpCon := webClient.NewHTTPConnector(config.MGM.SessionSecret, jMgr, sim, uMgr, mailer, logger)
	sockCon := webClient.NewWebsocketConnector(httpCon, sessionListenerChan, logger)

	r := mux.NewRouter()
	r.HandleFunc("/ws", sockCon.WebsocketHandler)
	r.HandleFunc("/auth", httpCon.ResumeHandler)
	r.HandleFunc("/auth/login", httpCon.LoginHandler)
	r.HandleFunc("/auth/logout", httpCon.LogoutHandler)
	r.HandleFunc("/auth/register", httpCon.RegisterHandler)
	r.HandleFunc("/auth/passwordToken", httpCon.PasswordTokenHandler)
	r.HandleFunc("/auth/passwordReset", httpCon.PasswordResetHandler)
	r.HandleFunc("/upload/{id}", httpCon.UploadHandler)

	http.Handle("/", r)
	logger.Info("Listening for clients on :%d", config.MGM.WebPort)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.MGM.WebPort), nil); err != nil {
		logger.Fatal("ListenAndServe:", err)
	}
}
