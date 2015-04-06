package main

import (
  "./go_modules/mgm"
  "./go_modules/simian"
  "fmt"
  "net/http"
  "log"
  "os"
  "encoding/json"
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
  //config := Configuration{}
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
    
  regionMgr := mgm.ClientManager{}
    
  // listen for opensim connections
  opensim := mgm.OpenSimListener{config.OpensimPort}
  go opensim.Listen()
  
  // listen for client connections
  fs := http.FileServer(http.Dir("dist"))
  http.Handle("/", fs)
  http.Handle("/ws", mgm.ClientWebsocketHandler{regionMgr})
  http.Handle("/auth/login", mgm.ClientAuthHandler{regionMgr})
  fmt.Println("Listening for clients on :" + config.WebPort)
  if err := http.ListenAndServe(":" + config.WebPort, nil); err != nil {
    log.Fatal("ListenAndServe:", err)
  }
}