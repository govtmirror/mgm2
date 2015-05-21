package mgm

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"
)

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

// ObjectType implements UserObject
func (u User) ObjectType() string {
	return "User"
}

// Serialize implements UserObject
func (u User) Serialize() []byte {
	data, _ := json.Marshal(u)
	return data
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

// ObjectType implements UserObject
func (u PendingUser) ObjectType() string {
	return "PendingUser"
}

// Serialize implements UserObject
func (u PendingUser) Serialize() []byte {
	data, _ := json.Marshal(u)
	return data
}
