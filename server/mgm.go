package main

import (
	"github.com/M-O-S-E-S/mgm/core"
	"github.com/M-O-S-E-S/mgm/email"
	"github.com/M-O-S-E-S/mgm/mysql"
	"github.com/M-O-S-E-S/mgm/simian"
	"github.com/M-O-S-E-S/mgm/webClient"
	//"github.com/M-O-S-E-S/mgm/opensim"
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

	cfgPtr := flag.String("config", "/opt/mgm/conf.gcfg", "path to config file")

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
	)
	//create our simian connector
	sim, err := simian.NewSimianConnector(config.MGM.SimianURL)
	if err != nil {
		logger.Error("Error instantiating Simian connection: ", err)
		return
	}

	//start new goroutine exiring old password tokens
	go expirePasswordTokens(db)

	//leave this out for now
	//os,_ := opensim.NewOpensimListener(config.OpensimPort, nil)

	jobNotifier := make(chan core.Job, 32)
	fileUpload := make(chan core.FileUpload, 32)

	//Hook up core processing...
	//regionManager := core.RegionManager{nil, db}
	sessionListener := make(chan core.UserSession, 64)
	core.UserManager(sessionListener, jobNotifier, db, sim, logger)
	core.JobManager(fileUpload, jobNotifier, config.MGM.LocalFileStorage, db, logger)

	httpCon := webClient.NewHttpConnector(config.MGM.SessionSecret, fileUpload, sim, db, mailer, logger)
	sockCon := webClient.NewWebsocketConnector(httpCon, sessionListener, logger)

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

func expirePasswordTokens(db *mysql.Database) {
	for {
		db.ExpirePasswordTokens()
		time.Sleep(60 * time.Minute)
	}
}
