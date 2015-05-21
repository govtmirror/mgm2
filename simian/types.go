package simian

import (
	"github.com/M-O-S-E-S/mgm/mgm"
	"github.com/satori/go.uuid"
)

type confirmRequest struct {
	Success bool
	Message string
}

type userRequest struct {
	Success bool
	Message string
	User    mgm.User
}

type usersRequest struct {
	Success bool
	Message string
	Users   []mgm.User
}

type Generic struct {
	OwnerID uuid.UUID
	Key     uuid.UUID
	Type    string
	Value   string
}
