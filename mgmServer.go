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

type Configuration struct {
  SimianUrl string
}

const (
  LINK_HOST = "127.0.0.1"
  LINK_PORT = "8000"
  CLIENT_PORT = "8080"
)

func main() {
  fmt.Println("Reading configuration file")
  file, _ := os.Open("conf.json")
  decoder := json.NewDecoder(file)
  config := Configuration{}
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
  opensim := mgm.OpenSimListener{Host:LINK_HOST, Port:LINK_PORT}
  go opensim.Listen()
  
  // listen for client connections
  fs := http.FileServer(http.Dir("/var/www/html/mgm"))
  http.Handle("/", fs)
  http.Handle("/ws", mgm.ClientHandler{regionMgr})
  fmt.Println("Listening for clients on 127.0.0.1:" + CLIENT_PORT)
  if err := http.ListenAndServe(":" + CLIENT_PORT, nil); err != nil {
    log.Fatal("ListenAndServe:", err)
  }
}