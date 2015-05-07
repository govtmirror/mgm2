package core

import (
  "github.com/satori/go.uuid"
  "encoding/json"
  "time"
)

type Opensim interface {
  Listen()
}

type User struct {
  UserID uuid.UUID
  Name string
  Email string
  AccessLevel uint8
  Suspended bool

  HomeLocation json.RawMessage  `json:"-"`
  LastLocation json.RawMessage `json:"-"`
  LLAbout json.RawMessage `json:"-"`
  LLInterests json.RawMessage `json:"-"`
  LLPackedAppearance json.RawMessage `json:"-"`
}

type PendingUser struct {
  Name string
  Email string
  Gender string
  PasswordHash string
  Registered time.Time
  Summary string
}

type Group struct {
  Name string
  Founder uuid.UUID
  ID uuid.UUID
  Members []uuid.UUID
  Roles []string
}

type Job struct {
  ID int
  Timestamp time.Time
  Type string
  User uuid.UUID
  Data string
}

type LoadIarTask struct {
  InventoryPath string
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

type Host struct {
  Address string
  Port uint           `json:"-"`
  Hostname string
  Regions []uuid.UUID
  Status string
  Slots uint
}

type ConfigOption struct {
  Region uuid.UUID
  Section string
  Item string
  Content string
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
  SlaveAddress string     `json:"-"`
  IsRunning bool
  Status string
  EstateName string
  
  frames chan int
}
