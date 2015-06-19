package mgm

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

//UserMessage is the struct for sending requests to and non-object responses from a user session
type UserMessage struct {
	MessageID   int
	MessageType string
	Message     json.RawMessage
}

// Load parses a json []byte onto itsself
func (ur *UserMessage) Load(msg []byte) {
	err := json.Unmarshal(msg, ur)
	if err != nil {
		ur.MessageType = err.Error()
	}
}

// ReadID parses an {ID: int} from the Message body
func (ur UserMessage) ReadID() (int64, error) {
	type id struct {
		ID int64
	}
	r := id{}
	err := json.Unmarshal(ur.Message, &r)
	if err != nil {
		return 0, err
	}
	return r.ID, nil
}

// ReadRegionID parses {RegionUUID: uuid.UUID} from the message body
func (ur UserMessage) ReadRegionID() (uuid.UUID, error) {
	type rid struct {
		RegionUUID uuid.UUID
	}
	r := rid{}
	err := json.Unmarshal(ur.Message, &r)
	if err != nil {
		return uuid.UUID{}, err
	}
	return r.RegionUUID, nil
}

// ReadPassword parses {UserID: uuid.UUID, Password: string} from the message body
func (ur UserMessage) ReadPassword() (uuid.UUID, string, error) {
	type pw struct {
		UserID   uuid.UUID
		Password string
	}
	p := pw{}
	err := json.Unmarshal(ur.Message, &p)
	if err != nil {
		return uuid.UUID{}, "", err
	}
	return p.UserID, p.Password, nil
}

// ReadAddress parses {Address: string} from the message body
func (ur UserMessage) ReadAddress() (string, error) {
	type pw struct {
		Address string
	}
	p := pw{}
	err := json.Unmarshal(ur.Message, &p)
	if err != nil {
		return "", err
	}
	return p.Address, nil
}

// ObjectType implements UserObject interface
func (ur UserMessage) ObjectType() string {
	return ur.MessageType
}

// Serialize implements UserObject
func (ur UserMessage) Serialize() []byte {
	data, _ := json.Marshal(ur)
	return data
}
