package core

import (
	"github.com/m-o-s-e-s/mgm/email"
	"github.com/satori/go.uuid"
)

// MgmConfig is a struct for parsing the gcfg main config file
type MgmConfig struct {
	MGM struct {
		MgmURL        string
		SimianURL     string
		SecretKey     string
		OpensimPort   string
		WebPort       int
		NodePort      int
		HubRegionUUID uuid.UUID
	}

	Web struct {
		Root        string
		Debug       bool
		Hostname    string
		FileStorage string
	}

	MySQL struct {
		Username string
		Password string
		Host     string
		Database string
	}

	Opensim struct {
		Username string
		Password string
		Host     string
		Database string
	}

	Email email.EmailConfig
}
