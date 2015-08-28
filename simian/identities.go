package simian

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/satori/go.uuid"
)

func (sc Connector) Auth(username string, password string) (bool, uuid.UUID, error) {
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
		return false, uuid.UUID{}, fmt.Errorf("Error communicating with simian: %v", err)
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

// EnableIdentity overwrites a given identity with Enabled=1
func (sc Connector) EnableIdentity(username string, identityType string, credential string, userID uuid.UUID) error {
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
		return fmt.Errorf("Error communicating with simian: %v", err)
	}

	return sc.confirmRequest(response)
}

// DisableIdentity overwrites a given identity with Enabled=0
func (sc Connector) DisableIdentity(username string, identityType string, credential string, userID uuid.UUID) error {
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
		return fmt.Errorf("Error communicating with simian: %v", err)
	}

	return sc.confirmRequest(response)
}

//InsertPasswordHash inserts a specified password hash into Simian
func (sc Connector) InsertPasswordHash(username string, credential string, userID uuid.UUID) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddIdentity"},
			"Identifier":    {username},
			"Type":          {"md5hash"},
			"Credential":    {credential},
			"UserID":        {userID.String()},
		})

	if err != nil {
		return fmt.Errorf("Error communicating with simian: %v", err)
	}

	return sc.confirmRequest(response)
}

// SetPassword hashes and inserts a plain password into Simian
func (sc Connector) SetPassword(userID uuid.UUID, password string) error {
	hasher := md5.New()
	hasher.Write([]byte(password))

	user, _, _ := sc.GetUserByID(userID)

	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddIdentity"},
			"Identifier":    {user.Name},
			"Type":          {"md5hash"},
			"Credential":    {"$1$" + hex.EncodeToString(hasher.Sum(nil))},
			"UserID":        {userID.String()},
		})

	if err != nil {
		return fmt.Errorf("Error communicating with simian: %v", err)
	}

	return sc.confirmRequest(response)
}

// ValidatePassword tests a given password against the users credential in Simian
func (sc Connector) ValidatePassword(userID uuid.UUID, password string) (bool, error) {
	user, exists, err := sc.GetUserByID(userID)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
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

// GetIdentities queries all given identities for a given user
func (sc Connector) GetIdentities(userID uuid.UUID) ([]core.Identity, error) {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetIdentities"},
			"UserID":        {userID.String()},
		})

	if err != nil {
		return nil, fmt.Errorf("Error communicating with simian: %v", err)
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
	return nil, fmt.Errorf("Error communicating with simian: %v", m.Message)
}

// IsNameTaken tests whether a given name is already present in Simian
func (sc Connector) IsNameTaken(name string) (bool, error) {
	_, exists, err := sc.GetUserByName(name)
	if err != nil {
		return true, err
	}
	if exists {
		return true, nil
	}
	return false, nil
}

// IsEmailTaken tests whether a given email is already present in simian
func (sc Connector) IsEmailTaken(email string) (bool, error) {
	_, exists, err := sc.GetUserByEmail(email)
	if err != nil {
		return true, err
	}
	if exists {
		return true, nil
	}
	return false, nil
}
