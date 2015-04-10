package main

import (
  "github.com/M-O-S-E-S/mgm2/core"
  "github.com/M-O-S-E-S/mgm2/mysql"
  "github.com/M-O-S-E-S/mgm2/simian"
  "github.com/M-O-S-E-S/mgm2/webClient"
  "github.com/M-O-S-E-S/mgm2/opensim"
  "fmt"
  "os"
  "encoding/json"
  "net/http"
  "log"
  "github.com/gorilla/mux"
)

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

func main() {
  config := MgmConfig{}
  
  fmt.Println("Reading configuration file")
  file, _ := os.Open("conf.json")
  decoder := json.NewDecoder(file)

  err := decoder.Decode(&config)
  if err != nil {
    fmt.Println("Error readig config file: ", err)
    return
  }
  
  db := mysql.NewDatabase(
    config.DBUsername, 
    config.DBPassword, 
    config.DBDatabase,
    config.DBHost, 
  )
  sim, _ := simian.NewSimianConnector(config.SimianUrl)
  os,_ := opensim.NewOpensimListener(config.OpensimPort, nil)
  
  
  //NewMGM(config MgmConfig, simian Simian, database Database, opensim Opensim)
  mgmCore, err := core.NewMGM(sim, db, os)
  if err != nil {
    fmt.Println("Error instantiating MGMCore ", err)
    return
  }
  fmt.Println(mgmCore)

  httpCon := webClient.NewHttpConnector(config.SessionSecret, sim)
  sockCon := webClient.NewWebsocketConnector(httpCon)
  
  r := mux.NewRouter()
  r.HandleFunc("/ws", sockCon.WebsocketHandler)
  r.HandleFunc("/auth", httpCon.ResumeHandler)
  r.HandleFunc("/auth/login", httpCon.LoginHandler)
  r.HandleFunc("/auth/logout", httpCon.LogoutHandler)
  r.HandleFunc("/auth/register", httpCon.RegisterHandler)
  r.HandleFunc("/auth/passwordToken", httpCon.PasswordTokenHandler)
  r.HandleFunc("/auth/passwordReset", httpCon.PasswordResetHandler)
  
  http.Handle("/", r)
  fmt.Println("Listening for clients on :" + config.WebPort)
  if err := http.ListenAndServe(":" + config.WebPort, nil); err != nil {
    log.Fatal("ListenAndServe:", err)
  }
}