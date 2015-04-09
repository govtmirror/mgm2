package simian

import (
  "github.com/satori/go.uuid"
)

type confirmRequest struct {
  Success bool
  Message string
}

type userRequest struct {
  Success bool
  Message string
  User User
}

type usersRequest struct {
  Success bool
  Message string
  Users []User
}

type User struct {
  AccessLevel int
  Email string
  UserID uuid.UUID
  Cap uuid.UUID
  Name string
  LLPackedAppearance string
  HomeLocation string
  LastLocation string
  LLInterests string
  LLAbout string
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
