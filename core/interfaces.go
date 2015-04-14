package core

import "github.com/satori/go.uuid"

type UserSession interface {
  SendUserAccount(User)
  SendUserRegion(Region)
  Read() ([]byte, bool)

  GetGuid() uuid.UUID
}

type UserConnector interface {
  GetUserByID(uuid.UUID) (User, error)
}

type Database interface {
  TestConnection() error
  GetRegionsFor(uuid.UUID) ([]Region, error)
}
