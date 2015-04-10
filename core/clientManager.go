package core

import (
  "fmt"
  "github.com/gorilla/websocket"
  "github.com/satori/go.uuid"
)

type Client struct {
  guid uuid.UUID
  ws *websocket.Conn
  send chan []byte
}

type ClientManager struct {
  authIn chan ClientAuth
  authTest chan ClientAuthRequest
  authDel chan ClientAuth
  regionMgr RegionManager
}

type ClientAuth struct {
  guid uuid.UUID 
  token uuid.UUID
  address string
}

type ClientAuthRequest struct {
  Client ClientAuth
  Callback chan <- bool
}

func (cm ClientManager) NewClient(ws *websocket.Conn) *Client {
  fmt.Println("New client constructed")
  client := &Client{send: make(chan []byte, 256), ws: ws}
  return client
}