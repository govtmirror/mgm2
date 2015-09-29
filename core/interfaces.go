package core

import "github.com/satori/go.uuid"
import "github.com/m-o-s-e-s/mgm/mgm"

// UserSession is the connection to the web client
type UserSession interface {
	Send(UserObject)
	SignalSuccess(int, string)
	SignalError(int, string)
	SignalProgress(int, string)
	Read(chan<- []byte)

	GetGUID() uuid.UUID
	GetAccessLevel() uint8
	GetClosingSignal() <-chan bool
}

// UserObject is an object that is transmittable to the client
type UserObject interface {
	Serialize() []byte
	ObjectType() string
}

// UserConnector is the connection to the user services
type UserConnector interface {
	GetUserByID(uuid.UUID) (mgm.User, bool, error)
	GetUserByEmail(email string) (mgm.User, bool, error)
	GetUserByName(name string) (mgm.User, bool, error)
	GetUsers() ([]mgm.User, error)
	GetGroups() ([]mgm.Group, error)

	CreateUserEntry(string, string) (uuid.UUID, error)
	CreateUserInventory(uuid.UUID, string) (bool, error)

	UpdateUser(string, string, uuid.UUID, int) error

	SetPassword(uuid.UUID, string) error
	ValidatePassword(uuid.UUID, string) (bool, error)
	Auth(username string, password string) (bool, uuid.UUID, error)
}
