package simian

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// GetUserByEmail retrieves a user identified by the given email
func (sc Connector) GetUserByEmail(email string) (mgm.User, bool, error) {
	m := mgm.User{}
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetUser"},
			"Email":         {email},
		})

	if err != nil {
		return m, false, fmt.Errorf("Error communicating with simian: %v", err)
	}

	var rq userRequest
	err = json.Unmarshal(response, &rq)
	if err != nil {
		return rq.User, false, err
	}
	if rq.Success {
		return rq.User, true, nil
	}
	return m, false, nil
}

// GetUserByName retrieves a user identified by the given name
func (sc Connector) GetUserByName(name string) (mgm.User, bool, error) {
	m := mgm.User{}
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetUser"},
			"Name":          {name},
		})

	if err != nil {
		return m, false, fmt.Errorf("Error communicating with simian: %v", err)
	}

	var rq userRequest
	err = json.Unmarshal(response, &rq)
	if err != nil {
		return m, false, err
	}
	if rq.Success {
		return rq.User, true, nil
	}
	return m, false, nil
}

// GetUserByID retrieves a user identified by the given uuid
func (sc Connector) GetUserByID(id uuid.UUID) (mgm.User, bool, error) {
	m := mgm.User{}
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetUser"},
			"UserID":        {id.String()},
		})

	var rq userRequest
	err = json.Unmarshal(response, &rq)
	if err != nil {
		return m, false, err
	}
	if rq.Success {
		return rq.User, true, nil
	}
	return m, false, nil
}

// GetUsers retrieves all users present in Simian
func (sc Connector) GetUsers() ([]mgm.User, error) {
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
		return nil, fmt.Errorf("Error communicating with simian: %v", m.Message)
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

// RemoveUser removes a user record from Simian
func (sc Connector) RemoveUser(userID uuid.UUID) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"RemoveUser"},
			"UserID":        {userID.String()},
		})

	if err != nil {
		return fmt.Errorf("Error communicating with simian: %v", err)
	}

	return sc.confirmRequest(response)
}

// SetUserLastLocation utility function to ensure users last location is set
func (sc Connector) SetUserLastLocation(userID uuid.UUID, uri string) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddUserData"},
			"UserID":        {userID.String()},
			"LastLocation":  {uri},
		})

	if err != nil {
		return fmt.Errorf("Error communicating with simian: %v", err)
	}

	return sc.confirmRequest(response)
}

// SetUserHomeLocation update a given users home location
func (sc Connector) SetUserHomeLocation(userID uuid.UUID, uri string) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddUserData"},
			"UserID":        {userID.String()},
			"HomeLocation":  {uri},
		})

	if err != nil {
		return fmt.Errorf("Error communicating with simian: %v", err)
	}

	return sc.confirmRequest(response)
}

// UpdateUser updates a user record
func (sc Connector) UpdateUser(name string, email string, userID uuid.UUID, level int) error {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"AddUser"},
			"UserID":        {userID.String()},
			"Email":         {email},
			"Name":          {name},
			"AccessLevel":   {fmt.Sprintf("%v", level)},
		})

	if err != nil {
		return fmt.Errorf("Error communicating with simian: %v", err)
	}

	return sc.confirmRequest(response)
}
