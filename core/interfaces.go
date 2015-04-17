package core

import "github.com/satori/go.uuid"

type UserSession interface {
  SendUser(User)
  SendPendingUser(PendingUser)
  SendRegion(Region)
  SendEstate(Estate)
  SendGroup(Group)
  SendHost(Host)
  SendConfig(ConfigOption)
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
  GetRegionsForUser(uuid.UUID) ([]Region, error)
  GetRegionsOnHost(string) ([]Region, error)
  GetRegions()([]Region, error)
  GetEstates()([]Estate, error)
  GetHosts()([]Host, error)

  GetPendingUsers() ([]PendingUser, error)

  GetDefaultConfigs()([]ConfigOption, error)
}

type Logger interface {
  Trace(format string, v ...interface{})
  Debug(format string, v ...interface{})
  Info(format string, v ...interface{})
  Warn(format string, v ...interface{})
  Error(format string, v ...interface{})
  Fatal(format string, v ...interface{})
}
