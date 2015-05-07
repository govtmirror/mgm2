package core

import "github.com/satori/go.uuid"

type UserSession interface {
  SendUser(int, User)
  SendPendingUser(int, PendingUser)
  SendRegion(int, Region)
  SendEstate(int, Estate)
  SendGroup(int, Group)
  SendHost(int, Host)
  SendConfig(int, ConfigOption)
  SendJob(int, Job)
  SignalSuccess(int, string)
  SignalError(int, string)
  Read() ([]byte, bool)

  GetGuid() uuid.UUID
  GetAccessLevel() uint8
}

type UserConnector interface {
  GetUserByID(uuid.UUID) (*User, error)
  GetUsers() ([]User, error)
  GetGroups() ([]Group, error)

  SetPassword(uuid.UUID, string) error
  ValidatePassword(uuid.UUID, string) (bool, error)
}

type Database interface {
  TestConnection() error
  GetRegionsForUser(uuid.UUID) ([]Region, error)
  GetJobsForUser(uuid.UUID) ([]Job, error)
  GetRegionsOnHost(string) ([]Region, error)
  GetRegions()([]Region, error)
  GetEstates()([]Estate, error)
  GetHosts()([]Host, error)

  GetPendingUsers() ([]PendingUser, error)

  GetDefaultConfigs()([]ConfigOption, error)
  GetConfigs(uuid.UUID)([]ConfigOption, error)

  CreateJob(string, uuid.UUID, string) (Job, error)
  CreateLoadIarJob(owner uuid.UUID, inventoryPath string) (Job, error)
}

type Logger interface {
  Trace(format string, v ...interface{})
  Debug(format string, v ...interface{})
  Info(format string, v ...interface{})
  Warn(format string, v ...interface{})
  Error(format string, v ...interface{})
  Fatal(format string, v ...interface{})
}
