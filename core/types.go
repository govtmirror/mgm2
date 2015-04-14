package core

import (
  "github.com/satori/go.uuid"
  "encoding/json"
)

type Opensim interface {
  Listen()
}

type User struct {
  UserID uuid.UUID
  Name string
  Email string
  AccessLevel uint8

  HomeLocation json.RawMessage  `json:"-"`
  LastLocation json.RawMessage `json:"-"`
  LLAbout json.RawMessage `json:"-"`
  LLInterests json.RawMessage `json:"-"`
  LLPackedAppearance json.RawMessage `json:"-"`
}

type Region struct {
  UUID uuid.UUID
  Name string
  Size uint
  HttpPort int
  ConsolePort int
  ConsoleUname uuid.UUID
  ConsolePass uuid.UUID
  LocX uint
  LocY uint
  ExternalAddress string
  SlaveAddress string
  IsRunning bool
  Status string
  EstateName string
  
  frames chan int
}
