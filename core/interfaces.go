package core

import "github.com/satori/go.uuid"

type UserSession interface {
  SendUserAccount(User)
  SendUserRegion(Region)
  Read() ([]byte, bool)

  GetGuid() uuid.UUID
  GetAccessLevel() uint8
}

type UserConnector interface {
  GetUserByID(uuid.UUID) (*User, error)
}

type Database interface {
  TestConnection() error
  GetRegionsFor(uuid.UUID) ([]Region, error)
}

type Logger interface {
  Trace(format string, v ...interface{})
  Debug(format string, v ...interface{})
  Info(format string, v ...interface{})
  Warn(format string, v ...interface{})
  Error(format string, v ...interface{})
  Fatal(format string, v ...interface{})
}
