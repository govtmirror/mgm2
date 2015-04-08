package mgm

import (
  "../simian"
  "fmt"
  "net/http"
  "log"
  "github.com/gorilla/mux"
  
)

type mgmRequest struct {
  request string
  listener chan<- []byte
}

type MgmConfig struct {
  SimianUrl string
  SessionSecret string
  OpensimPort string
  WebPort string
  DBUsername string
  DBPassword string
  DBHost string
  DBDatabase string
}

type mgmCore struct{
  requests chan mgmRequest
  clientMgr clientManager
  config MgmConfig
}

var mgmInstance *mgmCore = nil

func NewMGM(config MgmConfig) (*mgmCore, error){
  if mgmInstance == nil {
    //Make sure that simian is happy
    fmt.Println("Initializing Simiangrid Connection")
    err := simian.Initialize(config.SimianUrl)
    if err != nil {
      return nil, err
    }
    
    //Instantiate our client manager with session keyMGM
    clientMgr := clientManager{}
    clientMgr.Initialize(config.SessionSecret)
    
    regionMgr := regionManager{}
    
    opensim := openSimListener{config.OpensimPort, regionMgr}
    
    mgmInstance = &mgmCore{make(chan mgmRequest, 256), clientMgr, config}
    
    //populate interna structures from database
    db := database{user: config.DBUsername, password: config.DBPassword, host: config.DBHost, database: config.DBDatabase}
    err = db.testConnection()
    if err != nil {
      return nil, err
    }
    
    go opensim.Listen()
    
    
  }
  return mgmInstance, nil
}

func (mgm *mgmCore) Listen(){
  fmt.Println("running")
  
  r := mux.NewRouter()
  r.HandleFunc("/ws", mgm.clientMgr.websocketHandler)
  r.HandleFunc("/auth", mgm.clientMgr.resumeHandler)
  r.HandleFunc("/auth/login", mgm.clientMgr.loginHandler)
  r.HandleFunc("/auth/logout", mgm.clientMgr.logoutHandler)
  r.HandleFunc("/auth/register", mgm.clientMgr.registerHandler)
  r.HandleFunc("/auth/passwordToken", mgm.clientMgr.passwordTokenHandler)
  r.HandleFunc("/auth/passwordReset", mgm.clientMgr.passwordResetHandler)
  
  http.Handle("/", r)
  fmt.Println("Listening for clients on :" + mgm.config.WebPort)
  if err := http.ListenAndServe(":" + mgm.config.WebPort, nil); err != nil {
    log.Fatal("ListenAndServe:", err)
  }
}