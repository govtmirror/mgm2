package core

import "github.com/satori/go.uuid"

// Identity is a simiangrid credential record
type Identity struct {
	Identifier string
	Credential string
	Type       string
	UserID     uuid.UUID
	Enabled    bool
}

// ServiceRequest is a callback template for MGM services
type ServiceRequest func(bool, string)
