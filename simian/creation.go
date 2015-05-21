package simian

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/satori/go.uuid"
)

func (sc simian) EmailIsRegistered(email string) (exists bool, err error) {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetUser"},
			"Email":         {email},
		})

	if err != nil {
		return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	var m confirmRequest
	err = json.Unmarshal(response, &m)
	if err != nil {
		return false, err
	}
	if m.Success {
		return true, nil
	}
	return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}

func (sc simian) CreateUserEntry(username string, email string) (uuid.UUID, error) {
	userID := uuid.NewV4()
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddUser"},
			"UserID":        {userID.String()},
			"Name":          {username},
			"Email":         {email},
			"AccessLevel":   {"0"},
		})

	if err != nil {
		return uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	var m confirmRequest
	err = json.Unmarshal(response, &m)
	if err != nil {
		return uuid.UUID{}, err
	}
	if m.Success {
		return userID, nil
	}
	return uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}

func (sc simian) CreateUserInventory(userID uuid.UUID, template string) (bool, error) {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddInventory"},
			"OwnerID":       {userID.String()},
			"AvatarType":    {template},
		})

	var m confirmRequest
	err = json.Unmarshal(response, &m)
	if err != nil {
		return false, err
	}
	if m.Success {
		return true, nil
	}
	return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}
