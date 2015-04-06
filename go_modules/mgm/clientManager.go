package mgm

import (
  "fmt"
  "github.com/gorilla/websocket"
  "github.com/gorilla/sessions"
  "github.com/satori/go.uuid"
)

type ClientManager struct {
  authIn chan clientAuth
  authTest chan clientAuthRequest
  authDel chan clientAuth
  store *sessions.CookieStore
}

type clientAuth struct {
  guid uuid.UUID 
  token uuid.UUID
  address string
}

type clientAuthRequest struct {
  client clientAuth
  callback chan <- bool
}

func (cm *ClientManager) NewClient(ws *websocket.Conn) *Client {
  fmt.Println("New client constructed")
  client := &Client{send: make(chan []byte, 256), ws: ws}
  return client
}

func (cm *ClientManager) Initialize(sessionKey string){
  cm.store = sessions.NewCookieStore([]byte(sessionKey))
}