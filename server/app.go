package main

import (
  "github.com/M-O-S-E-S/mgm2/core"
  "github.com/M-O-S-E-S/mgm2/mysql"
  "github.com/M-O-S-E-S/mgm2/simian"
  "github.com/M-O-S-E-S/mgm2/webClient"
  "github.com/M-O-S-E-S/mgm2/email"
  //"github.com/M-O-S-E-S/mgm2/opensim"
  "net/http"
  "github.com/gorilla/mux"
  "code.google.com/p/gcfg"
  "github.com/jcelliott/lumber"
  "time"
)

type MgmConfig struct {
  MGM struct {
    SimianUrl string
    SessionSecret string
    OpensimPort string
    WebPort string
    PublicHostname string
  }

  MySQL struct {
    Username string
    Password string
    Host string
    Database string
  }

  Email email.EmailConfig
}

func main() {
  //instantiate our logger
  logger := lumber.NewConsoleLogger(lumber.DEBUG)

  //read configuration file
  config := MgmConfig{}
  err := gcfg.ReadFileInto(&config, "conf.gcfg")
  if err != nil {
    logger.Fatal("Error reading config file: %v", err)
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
  sim, _ := simian.NewSimianConnector(config.MGM.SimianUrl)

  //start new goroutine exiring old password tokens
  go ExpirePasswordTokens(db)
  
  //leave this out for now
  //os,_ := opensim.NewOpensimListener(config.OpensimPort, nil)
  
  
  //Hook up core processing...
  //regionManager := core.RegionManager{nil, db}
  sessionListener := make(chan core.UserSession, 64) 
  core.UserManager(sessionListener, db, sim, logger)

  httpCon := webClient.NewHttpConnector(config.MGM.SessionSecret, sim, db, mailer, logger)
  sockCon := webClient.NewWebsocketConnector(httpCon, sessionListener, logger)
  
  r := mux.NewRouter()
  r.HandleFunc("/ws", sockCon.WebsocketHandler)
  r.HandleFunc("/auth", httpCon.ResumeHandler)
  r.HandleFunc("/auth/login", httpCon.LoginHandler)
  r.HandleFunc("/auth/logout", httpCon.LogoutHandler)
  r.HandleFunc("/auth/register", httpCon.RegisterHandler)
  r.HandleFunc("/auth/passwordToken", httpCon.PasswordTokenHandler)
  r.HandleFunc("/auth/passwordReset", httpCon.PasswordResetHandler)
  
  http.Handle("/", r)
  logger.Info("Listening for clients on :%v", config.MGM.WebPort)
  if err := http.ListenAndServe(":" + config.MGM.WebPort, nil); err != nil {
    logger.Fatal("ListenAndServe:", err)
  }
}

func ExpirePasswordTokens(db *mysql.Database){
  for {
    db.ExpirePasswordTokens()
    time.Sleep(60 * time.Minute)
  }
}
