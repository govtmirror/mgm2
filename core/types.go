package core

import (
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// Identity is a simiangrid credential record
type Identity struct {
	Identifier string
	Credential string
	Type       string
	UserID     uuid.UUID
	Enabled    bool
}

// SessionLookup is a struct for session lookup tables
type SessionLookup struct {
	JobLink      chan mgm.Job
	HostStatLink chan mgm.HostStat
	HostLink     chan mgm.Host
	AccessLevel  uint8
}

// ServiceRequest is a callback template for MGM services
type ServiceRequest func(bool, string)
