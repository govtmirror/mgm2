package core

import "github.com/satori/go.uuid"
import "github.com/M-O-S-E-S/mgm/mgm"

// UserSession is the connection to the web client
type UserSession interface {
	GetSend() chan<- UserObject
	SignalSuccess(int, string)
	SignalError(int, string)
	Read(chan<- []byte, chan<- bool)

	GetGUID() uuid.UUID
	GetAccessLevel() uint8
}

// UserObject is an object that is transmittable to the client
type UserObject interface {
	Serialize() []byte
	ObjectType() string
}

// UserConnector is the connection to the user services
type UserConnector interface {
	GetUserByID(uuid.UUID) (*mgm.User, error)
	GetUsers() ([]mgm.User, error)
	GetGroups() ([]mgm.Group, error)

	SetPassword(uuid.UUID, string) error
	ValidatePassword(uuid.UUID, string) (bool, error)
}

// Database is the connection to the persistant storage
type Database interface {
	TestConnection() error
	GetRegionsForUser(uuid.UUID) ([]mgm.Region, error)
	GetJobsForUser(uuid.UUID) ([]mgm.Job, error)
	GetRegionsOnHost(mgm.Host) ([]mgm.Region, error)
	GetRegions() ([]mgm.Region, error)
	GetEstates() ([]mgm.Estate, error)
	GetHosts() ([]mgm.Host, error)
	PlaceHostOnline(uint) (mgm.Host, error)
	PlaceHostOffline(uint) (mgm.Host, error)
	GetHostByAddress(string) (mgm.Host, error)

	GetPendingUsers() ([]mgm.PendingUser, error)

	GetDefaultConfigs() ([]ConfigOption, error)
	GetConfigs(uuid.UUID) ([]ConfigOption, error)

	CreateJob(string, uuid.UUID, string) (mgm.Job, error)
	CreateLoadIarJob(uuid.UUID, string) (mgm.Job, error)
	UpdateJob(mgm.Job) error
	DeleteJob(mgm.Job) error

	GetJobByID(int) (mgm.Job, error)

	CreatePasswordResetToken(uuid.UUID) (uuid.UUID, error)
	ValidatePasswordToken(uuid.UUID, uuid.UUID) (bool, error)
	ScrubPasswordToken(uuid.UUID) error
	IsEmailUnique(string) (bool, error)
	IsNameUnique(string) (bool, error)
	AddPendingUser(name string, email string, template string, password string, summary string) error
}

// Logger is the system logging interface
type Logger interface {
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Fatal(format string, v ...interface{})
}
