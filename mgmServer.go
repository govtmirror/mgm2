package main

import (
  "./go_modules/mgm"
  "fmt"
  "net/http"
  "log"
  "os"
  "encoding/json"
  "github.com/gorilla/mux"
  
)

//global config object
type config struct {
  WebPort string
}

func main() {
  mgmConfig := mgm.MgmConfig{}
  conf := config{}
  
  fmt.Println("Reading configuration file")
  file, _ := os.Open("conf.json")
  decoder := json.NewDecoder(file)

  err := decoder.Decode(
    &struct{ 
      *mgm.MgmConfig
      *config
    }{&mgmConfig,&conf})
  if err != nil {
    fmt.Println("Error readig config file: ", err)
    return
  }
  
  mgmCore, err := mgm.NewMGM(mgmConfig)
  if err != nil {
    fmt.Println("Error instantiating MGMCore ", err)
    return
  }
  
  fmt.Println(mgmCore)
  
  fmt.Println("running")
  
  r := mux.NewRouter()
  r.HandleFunc("/ws", mgmCore.ClientMgr.WebsocketHandler)
  r.HandleFunc("/auth", mgmCore.ClientMgr.ResumeHandler)
  r.HandleFunc("/auth/login", mgmCore.ClientMgr.LoginHandler)
  r.HandleFunc("/auth/logout", mgmCore.ClientMgr.LogoutHandler)
  r.HandleFunc("/auth/register", mgmCore.ClientMgr.RegisterHandler)
  r.HandleFunc("/auth/passwordToken", mgmCore.ClientMgr.PasswordTokenHandler)
  r.HandleFunc("/auth/passwordReset", mgmCore.ClientMgr.PasswordResetHandler)
  
  http.Handle("/", r)
  fmt.Println("Listening for clients on :" + conf.WebPort)
  if err := http.ListenAndServe(":" + conf.WebPort, nil); err != nil {
    log.Fatal("ListenAndServe:", err)
  }
}