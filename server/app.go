package main

import (
  "github.com/M-O-S-E-S/mgm2/core"
  "github.com/M-O-S-E-S/mgm2/mysql"
  "github.com/M-O-S-E-S/mgm2/simian"
  "github.com/M-O-S-E-S/mgm2/webClient"
  "github.com/M-O-S-E-S/mgm2/email"
  //"github.com/M-O-S-E-S/mgm2/opensim"
  "fmt"
  "net/http"
  "github.com/gorilla/mux"
  "code.google.com/p/gcfg"
  "github.com/jcelliott/lumber"
)

type MgmConfig struct {
  MGM struct {
    SimianUrl string
    SessionSecret string
    OpensimPort string
    WebPort string
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
  config := MgmConfig{}
  err := gcfg.ReadFileInto(&config, "conf.gcfg")
  
  logger := lumber.NewConsoleLogger(lumber.DEBUG)

  //fmt.Println("Reading configuration file")
  //file, _ := os.Open("conf.json")
  //decoder := json.NewDecoder(file)

  //err := decoder.Decode(&config)
  if err != nil {
    fmt.Println("Error reading config file: ", err)
    return
  }

  mailer := email.NewClientMailer(config.Email)

  db := mysql.NewDatabase(
    config.MySQL.Username,
    config.MySQL.Password,
    config.MySQL.Database,
    config.MySQL.Host,
  )
  sim, _ := simian.NewSimianConnector(config.MGM.SimianUrl)
  
  //leave this out for now
  //os,_ := opensim.NewOpensimListener(config.OpensimPort, nil)
  
  
  //Hook up core processing...
  //regionManager := core.RegionManager{nil, db}
  sessionListener := make(chan core.UserSession, 64) 
  core.UserManager(sessionListener, db, sim, logger)

  httpCon := webClient.NewHttpConnector(config.MGM.SessionSecret, sim, logger)
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
