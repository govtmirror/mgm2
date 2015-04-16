package webClient

import (
  "encoding/json"
  "github.com/satori/go.uuid"
  "github.com/gorilla/websocket"
  "github.com/M-O-S-E-S/mgm2/core"
)

type client struct {
  ws *websocket.Conn
  toClient chan []byte
  fromClient chan []byte
  guid uuid.UUID
  userLevel uint8
}

func (c client) SendUser(account core.User){
  resp := clientResponse{ "AccountUpdate", account}
  data, err := json.Marshal(resp)
  if err == nil {
    c.toClient <- data
  }
}

func (c client) SendRegion(region core.Region){
  resp := clientResponse{ "RegionUpdate", region}
  data, err := json.Marshal(resp)
  if err == nil {
    c.toClient <- data
  }
}

func (c client) SendEstate(estate core.Estate){
  resp := clientResponse{ "EstateUpdate", estate}
  data, err := json.Marshal(resp)
  if err == nil {
    c.toClient <- data
  }
}

func (c client) GetGuid() uuid.UUID {
  return c.guid
}

func (c client) GetAccessLevel() uint8 {
  return c.userLevel
}

func (c client) Read() ([]byte, bool){
  data, more := <- c.fromClient
  return data, more
}
