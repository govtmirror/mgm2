package main

import (
	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/jobManager"
	"github.com/m-o-s-e-s/mgm/core/nodeManager"
	"github.com/m-o-s-e-s/mgm/core/regionManager"
	"github.com/m-o-s-e-s/mgm/core/sessionManager"
	"github.com/m-o-s-e-s/mgm/email"
	"github.com/m-o-s-e-s/mgm/mysql"
	"github.com/m-o-s-e-s/mgm/simian"
	"github.com/m-o-s-e-s/mgm/webClient"
	//"github.com/m-o-s-e-s/mgm/opensim"
	"flag"
	"net/http"
	"time"

	"code.google.com/p/gcfg"
	"github.com/gorilla/mux"
	"github.com/jcelliott/lumber"
)

type mgmConfig struct {
	MGM struct {
		SimianURL        string
		SessionSecret    string
		OpensimPort      string
		WebPort          string
		NodePort         string
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
	logger := lumber.NewConsoleLogger(lumber.DEBUG)

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
	db := mysql.NewDatabase(
		config.MySQL.Username,
		config.MySQL.Password,
		config.MySQL.Database,
		config.MySQL.Host,
		logger,
	)
	//create our simian connector
	sim, err := simian.NewConnector(config.MGM.SimianURL)
	if err != nil {
		logger.Error("Error instantiating Simian connection: ", err)
		return
	}

	//start new goroutine exiring old password tokens
	go expirePasswordTokens(db)

	//leave this out for now
	//os,_ := opensim.NewOpensimListener(config.OpensimPort, nil)

	//Hook up core processing...
	jMgr := jobManager.NewJobManager(config.MGM.LocalFileStorage, db, logger)
	nMgr := nodeManager.NewNodeManager(config.MGM.NodePort, db, logger)
	rMgr := regionManager.NewRegionManager(nMgr, db, logger)
	sessionListenerChan := make(chan core.UserSession, 64)

	_ = sessionManager.NewSessionManager(sessionListenerChan, jMgr, nMgr, rMgr, db, sim, logger)

	httpCon := webClient.NewHTTPConnector(config.MGM.SessionSecret, jMgr, sim, db, mailer, logger)
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
	logger.Info("Listening for clients on :%v", config.MGM.WebPort)
	if err := http.ListenAndServe(":"+config.MGM.WebPort, nil); err != nil {
		logger.Fatal("ListenAndServe:", err)
	}
}

func expirePasswordTokens(db mysql.Database) {
	for {
		db.ExpirePasswordTokens()
		time.Sleep(60 * time.Minute)
	}
}
