package core

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"
)

// Opensim is the interface to the Opensimulator Region
type Opensim interface {
	Listen()
}

// User is the user record
type User struct {
	UserID      uuid.UUID
	Name        string
	Email       string
	AccessLevel uint8
	Suspended   bool

	HomeLocation       json.RawMessage `json:"-"`
	LastLocation       json.RawMessage `json:"-"`
	LLAbout            json.RawMessage `json:"-"`
	LLInterests        json.RawMessage `json:"-"`
	LLPackedAppearance json.RawMessage `json:"-"`
}

// PendingUser is a user who has applied, but has not been approved or denied
type PendingUser struct {
	Name         string
	Email        string
	Gender       string
	PasswordHash string
	Registered   time.Time
	Summary      string
}

// Group is an Opensim group record
type Group struct {
	Name    string
	Founder uuid.UUID
	ID      uuid.UUID
	Members []uuid.UUID
	Roles   []string
}

// Job is a record for long-running user tasks in MGM
type Job struct {
	ID        int
	Timestamp time.Time
	Type      string
	User      uuid.UUID
	Data      string
}

// LoadIarJob is the data field for jobs that are of type load_iar
type LoadIarJob struct {
	InventoryPath string
	Filename      uuid.UUID
	Status        string
}

// Estate is an opensim estate record
type Estate struct {
	Name     string
	ID       uint
	Owner    uuid.UUID
	Managers []uuid.UUID
	Regions  []uuid.UUID
}

// Identity is a simiangrid credential record
type Identity struct {
	Identifier string
	Credential string
	Type       string
	UserID     uuid.UUID
	Enabled    bool
}

// Host is an MGM server record
type Host struct {
	ID       uint
	Address  string
	Port     uint `json:"-"`
	Hostname string
	Regions  []uuid.UUID
	Status   string
	Slots    uint
}

// ConfigOption is an opensim.ini configuration line record
type ConfigOption struct {
	Region  uuid.UUID
	Section string
	Item    string
	Content string
}

// Region is an opensim region record
type Region struct {
	UUID            uuid.UUID
	Name            string
	Size            uint
	HTTPPort        int       `json:"-"`
	ConsolePort     int       `json:"-"`
	ConsoleUname    uuid.UUID `json:"-"`
	ConsolePass     uuid.UUID `json:"-"`
	LocX            uint
	LocY            uint
	ExternalAddress string `json:"-"`
	SlaveAddress    string `json:"-"`
	IsRunning       bool
	Status          string
	EstateName      string

	frames chan int
}
