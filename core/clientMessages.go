package core

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

type UserRequest struct {
	MessageID   int
	MessageType string
	Message     json.RawMessage
}

func (ur UserRequest) Load(msg []byte) {
	err := json.Unmarshal(msg, ur)
	if err != nil {
		ur.MessageType = err.Error()
	}
}

func (ur UserRequest) ReadID() (int, error) {
	type id struct {
		ID int
	}
	r := id{}
	err := json.Unmarshal(ur.Message, &r)
	if err != nil {
		return 0, err
	}
	return r.ID, nil
}

func (ur UserRequest) ReadRegionID() (uuid.UUID, error) {
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

func (ur UserRequest) ReadPassword() (uuid.UUID, string, error) {
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
