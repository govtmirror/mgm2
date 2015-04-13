package simian

import (
  "github.com/satori/go.uuid"
  "github.com/M-O-S-E-S/mgm2/core"
)

type confirmRequest struct {
  Success bool
  Message string
}

type userRequest struct {
  Success bool
  Message string
  User core.User
}

type usersRequest struct {
  Success bool
  Message string
  Users []core.User
}

type Identity struct {
  Identifier string
  Credential uuid.UUID
  Type string
  UserID uuid.UUID
  Enabled bool
}

type Group struct {
  OwnerID uuid.UUID
  Key string
  Value string
  GroupID uuid.UUID
}

type Generic struct {
  OwnerID uuid.UUID
  Key uuid.UUID
  Type string
  Value string
}
