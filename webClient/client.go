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
  logger Logger
}

func (c client) SendUser(req int, account core.User){
  resp := clientResponse{0, "UserUpdate", account}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}
func (c client) SendPendingUser(req int, account core.PendingUser){
  resp := clientResponse{0, "PendingUserUpdate", account}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SendRegion(req int, region core.Region){
  resp := clientResponse{req, "RegionUpdate", region}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SendEstate(req int, estate core.Estate){
  resp := clientResponse{req, "EstateUpdate", estate}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SendGroup(req int, group core.Group){
  resp := clientResponse{req, "GroupUpdate", group}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SendConfig(req int, cfg core.ConfigOption){
  resp := clientResponse{req, "ConfigUpdate", cfg}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SendHost(req int, host core.Host){
  if host.Status == "" {
    host.Status = "{}"
  }
  resp := clientResponse{req, "HostUpdate", host}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SignalSuccess(req int){
  resp := clientResponse{req, "Success", ""}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func(c client) writeData(data []byte){
  defer func() {
    if x := recover(); x != nil {
      c.logger.Info("Attempt to write to closed client channel")
    }
  }()
  c.toClient <- data
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
