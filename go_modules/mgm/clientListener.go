package mgm

import (
  "github.com/gorilla/websocket"
  "fmt"
  "net/http"
  "encoding/json"
  "../../go_modules/simian"
)

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type ClientWebsocketHandler struct {
  ClientMgr ClientManager
}

func (ch ClientWebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  fmt.Println("New connection on ws")
  ws, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  c := ch.ClientMgr.NewClient(ws)
  c.process()
}


type ClientAuth struct {
  Username string
  Password string
}

type ClientAuthHandler struct {
  ClientMgr ClientManager
}

func (ch ClientAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  fmt.Println("New Auth connection")
  decoder := json.NewDecoder(r.Body)
  var t ClientAuth   
  err := decoder.Decode(&t)
  if err != nil {
    fmt.Println("Invalid auth request")
    return
  }
  
  sim, _ := simian.Instance()
  uuid,err := sim.Auth(t.Username,t.Password);
  if err != nil {
    fmt.Println(err)
  }
  fmt.Println(uuid)
  
}