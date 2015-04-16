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

type Group struct {
  Name string
  Founder uuid.UUID
  ID uuid.UUID
  Members []uuid.UUID
  Roles []string
}

type Estate struct {
  Name string
  ID uint
  Owner uuid.UUID
  Managers []uuid.UUID
  Regions []uuid.UUID
}

type Identity struct {
  Identifier string
  Credential string
  Type string
  UserID uuid.UUID
  Enabled bool
}

type Region struct {
  UUID uuid.UUID
  Name string
  Size uint
  HttpPort int            `json:"-"`
  ConsolePort int         `json:"-"`
  ConsoleUname uuid.UUID  `json:"-"`
  ConsolePass uuid.UUID   `json:"-"`
  LocX uint
  LocY uint
  ExternalAddress string  `json:"-"`
  SlaveAddress string
  IsRunning bool
  Status string
  EstateName string
  
  frames chan int
}
