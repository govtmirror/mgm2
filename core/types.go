package core

import "github.com/satori/go.uuid"

// LoadIarJob is the data field for jobs that are of type load_iar
type LoadIarJob struct {
	InventoryPath string
	Filename      uuid.UUID
	Status        string
}

// Identity is a simiangrid credential record
type Identity struct {
	Identifier string
	Credential string
	Type       string
	UserID     uuid.UUID
	Enabled    bool
}

// ConfigOption is an opensim.ini configuration line record
type ConfigOption struct {
	Region  uuid.UUID
	Section string
	Item    string
	Content string
}
