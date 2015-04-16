package core

import "github.com/satori/go.uuid"

type UserSession interface {
  SendUser(User)
  SendRegion(Region)
  SendEstate(Estate)
  SendGroup(Group)
  Read() ([]byte, bool)

  GetGuid() uuid.UUID
  GetAccessLevel() uint8
}

type UserConnector interface {
  GetUserByID(uuid.UUID) (*User, error)
  GetUsers() ([]User, error)

  GetGroups() ([]Group, error)
}

type Database interface {
  TestConnection() error
  GetRegionsFor(uuid.UUID) ([]Region, error)
  GetAllRegions()([]Region, error)

  GetEstates()([]Estate, error)
}

type Logger interface {
  Trace(format string, v ...interface{})
  Debug(format string, v ...interface{})
  Info(format string, v ...interface{})
  Warn(format string, v ...interface{})
  Error(format string, v ...interface{})
  Fatal(format string, v ...interface{})
}
