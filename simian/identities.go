package simian

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/M-O-S-E-S/mgm/core"
	"github.com/satori/go.uuid"
)

func (sc simian) Auth(username string, password string) (bool, uuid.UUID, error) {
	hasher := md5.New()
	hasher.Write([]byte(password))

	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AuthorizeIdentity"},
			"Identifier":    {username},
			"Credential":    {hex.EncodeToString(hasher.Sum(nil))},
			"Type":          {"md5hash"},
		})

	if err != nil {
		if err.Error() == "Missing identity or invalid credentials" {
			return false, uuid.UUID{}, nil
		}
		return false, uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	type tmpStruct struct {
		Success bool
		Message string
		UserID  string
	}
	var m tmpStruct
	err = json.Unmarshal(response, &m)
	if err != nil {
		return false, uuid.UUID{}, err
	}
	if m.Success {
		userID, _ := uuid.FromString(m.UserID)
		return true, userID, nil
	}
	return false, uuid.UUID{}, nil
}

func (sc simian) EnableIdentity(username string, identityType string, credential string, userID uuid.UUID) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddIdentity"},
			"Identifier":    {username},
			"Type":          {identityType},
			"Credential":    {credential},
			"UserID":        {userID.String()},
			"Enabled":       {"1"},
		})

	if err != nil {
		return &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	return sc.confirmRequest(response)
}

func (sc simian) DisableIdentity(username string, identityType string, credential string, userID uuid.UUID) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddIdentity"},
			"Identifier":    {username},
			"Type":          {identityType},
			"Credential":    {credential},
			"UserID":        {userID.String()},
			"Enabled":       {"0"},
		})

	if err != nil {
		return &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	return sc.confirmRequest(response)
}

func (sc simian) InsertPasswordHash(username string, credential string, userID uuid.UUID) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddIdentity"},
			"Identifier":    {username},
			"Type":          {"md5hash"},
			"Credential":    {credential},
			"UserID":        {userID.String()},
		})

	if err != nil {
		return &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	return sc.confirmRequest(response)
}

func (sc simian) SetPassword(userID uuid.UUID, password string) error {
	hasher := md5.New()
	hasher.Write([]byte(password))

	user, _ := sc.GetUserByID(userID)

	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddIdentity"},
			"Identifier":    {user.Name},
			"Type":          {"md5hash"},
			"Credential":    {"$1$" + hex.EncodeToString(hasher.Sum(nil))},
			"UserID":        {userID.String()},
		})

	if err != nil {
		return &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	return sc.confirmRequest(response)
}

func (sc simian) ValidatePassword(userID uuid.UUID, password string) (bool, error) {
	user, err := sc.GetUserByID(userID)
	if err != nil {
		return false, err
	}
	valid, uid, err := sc.Auth(user.Name, password)
	if err != nil {
		return false, err
	}
	if valid {
		if uid == userID {
			return true, nil
		}
	}
	return false, nil
}

func (sc simian) GetIdentities(userID uuid.UUID) ([]core.Identity, error) {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetIdentities"},
			"UserID":        {userID.String()},
		})

	if err != nil {
		return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	type req struct {
		Success    bool
		Message    string
		Identities []core.Identity
	}
	var m req
	err = json.Unmarshal(response, &m)
	if err != nil {
		return nil, err
	}
	if m.Success {
		return m.Identities, nil
	}
	return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}
