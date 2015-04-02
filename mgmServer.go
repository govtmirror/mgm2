package main

import (
  "./go_modules/mgm"
  "fmt"
  "net/http"
  "log"
)

const (
    LINK_HOST = "127.0.0.1"
    LINK_PORT = "8000"
    CLIENT_PORT = "8080"
)

func main() {
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