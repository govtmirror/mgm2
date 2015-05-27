package simian

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

func (sc simian) GetUserByEmail(email string) (mgm.User, error) {
	m := mgm.User{}
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetUser"},
			"Email":         {email},
		})

	if err != nil {
		return m, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	var rq userRequest
	err = json.Unmarshal(response, &rq)
	if err != nil {
		return rq.User, err
	}
	if rq.Success {
		return rq.User, nil
	}
	return m, errors.New("Could not find user in simian")
}

func (sc simian) GetUserByName(name string) (mgm.User, error) {
	m := mgm.User{}
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetUser"},
			"Name":          {name},
		})

	if err != nil {
		return m, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	var rq userRequest
	err = json.Unmarshal(response, &rq)
	if err != nil {
		return m, err
	}
	if rq.Success {
		return rq.User, nil
	}
	return m, errors.New("Could not find user in simian")
}

func (sc simian) GetUserByID(id uuid.UUID) (mgm.User, error) {
	m := mgm.User{}
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetUser"},
			"UserID":        {id.String()},
		})

	var rq userRequest
	err = json.Unmarshal(response, &rq)
	if err != nil {
		return m, err
	}
	if rq.Success {
		return rq.User, nil
	}
	return m, errors.New("Cannot locate user in simian")
}

func (sc simian) GetUsers() ([]mgm.User, error) {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetUsers"},
			"NameQuery":     {""},
		})

	var m usersRequest
	err = json.Unmarshal(response, &m)
	if err != nil {
		return nil, err
	}
	if !m.Success {
		return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
	}
	//lookup suspension status for each user
	users := m.Users
	for idx, user := range users {
		ids, err := sc.GetIdentities(user.UserID)
		if err != nil {
			continue
		}
		isActive := false
		for _, id := range ids {
			if id.Enabled {
				isActive = true
			}
		}
		if !isActive {
			users[idx].Suspended = true
		}
	}
	return users, nil
}

func (sc simian) RemoveUser(userID uuid.UUID) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"RemoveUser"},
			"UserID":        {userID.String()},
		})

	if err != nil {
		return &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	return sc.confirmRequest(response)
}

func (sc simian) SetUserLastLocation(userID uuid.UUID, uri string) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddUserData"},
			"UserID":        {userID.String()},
			"LastLocation":  {uri},
		})

	if err != nil {
		return &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	return sc.confirmRequest(response)
}

func (sc simian) SetUserHomeLocation(userID uuid.UUID, uri string) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddUserData"},
			"UserID":        {userID.String()},
			"HomeLocation":  {uri},
		})

	if err != nil {
		return &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	return sc.confirmRequest(response)
}

func (sc simian) UpdateUser(name string, email string, userID uuid.UUID, level int) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddUser"},
			"UserID":        {userID.String()},
			"Email":         {email},
			"Name":          {name},
			"AccessLevel":   {fmt.Sprintf("%v", level)},
		})

	if err != nil {
		return &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	return sc.confirmRequest(response)
}
