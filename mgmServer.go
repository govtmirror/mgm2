package main

import (
  "./go_modules/mgm"
  "./go_modules/simian"
  "fmt"
  "net/http"
  "log"
  "os"
  "encoding/json"
  "github.com/gorilla/mux"
  
)

//global config object
var config = struct {
  SimianUrl string
  SessionSecret string
  WebPort string
  OpensimPort string
}{}

func main() {
  
  fmt.Println("Reading configuration file")
  file, _ := os.Open("conf.json")
  decoder := json.NewDecoder(file)

  err := decoder.Decode(&config)
  if err != nil {
    fmt.Println("Error readig config file: ", err)
    return
  }
  
  fmt.Println("Initializing Simiangrid Connection")
  err = simian.Initialize(config.SimianUrl)
  if err != nil {
    fmt.Println("Error initializing simiangrid: ", err)
    return
  }
  
  fmt.Println("running")
    
  clientMgr := mgm.ClientManager{}
  clientMgr.Initialize(config.SessionSecret)
    
  // listen for opensim connections
  opensim := mgm.OpenSimListener{config.OpensimPort}
  go opensim.Listen()
  
  r := mux.NewRouter()
  r.HandleFunc("/ws", clientMgr.WebsocketHandler)
  r.HandleFunc("/auth", clientMgr.ResumeHandler)
  r.HandleFunc("/auth/login", clientMgr.LoginHandler)
  r.HandleFunc("/auth/logout", clientMgr.LogoutHandler)
  r.HandleFunc("/auth/register", clientMgr.RegisterHandler)
  r.HandleFunc("/auth/passwordToken", clientMgr.PasswordTokenHandler)
  r.HandleFunc("/auth/passwordReset", clientMgr.PasswordResetHandler)
  
  http.Handle("/", r)
  fmt.Println("Listening for clients on :" + config.WebPort)
  if err := http.ListenAndServe(":" + config.WebPort, nil); err != nil {
    log.Fatal("ListenAndServe:", err)
  }
}