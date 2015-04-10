package core

import "github.com/satori/go.uuid"

type Simian interface {
}

type Database interface {
  TestConnection() error
  GetAllRegions() error
}

type Opensim interface {
  Listen()
}

type User struct {
  guid uuid.UUID
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
  
  frames chan int
}
