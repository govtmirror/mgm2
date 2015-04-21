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

func (c client) SendUser(account core.User){
  resp := clientResponse{ "UserUpdate", account}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}
func (c client) SendPendingUser(account core.PendingUser){
  resp := clientResponse{ "PendingUserUpdate", account}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SendRegion(region core.Region){
  resp := clientResponse{ "RegionUpdate", region}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SendEstate(estate core.Estate){
  resp := clientResponse{ "EstateUpdate", estate}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SendGroup(group core.Group){
  resp := clientResponse{ "GroupUpdate", group}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SendConfig(cfg core.ConfigOption){
  resp := clientResponse{ "ConfigUpdate", cfg}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SendHost(host core.Host){
  if host.Status == "" {
    host.Status = "{}"
  }
  resp := clientResponse{ "HostUpdate", host}
  data, err := json.Marshal(resp)
  if err == nil {
    c.writeData(data)
  }
}

func (c client) SignalSyncComplete(){
  resp := clientResponse{ "SyncComplete", nil}
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
