package user

import (
	"errors"
	"fmt"
	"strings"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// Auth performs user lookup and authentication steps
func (um Manager) Auth(username string, password string) (mgm.User, bool) {
	um.uMutex.Lock()
	defer um.uMutex.Unlock()

	//make sure user exists
	var user mgm.User
	found := false
	for _, u := range um.users {
		if strings.EqualFold(u.Name, username) {
			user = u
			found = true
		}
	}

	if found == false {
		um.log.Info("User %v does not exist", username)
		return user, false
	}

	//test user password
	valid, guid, err := um.conn.Auth(username, password)
	if err != nil {
		um.log.Error(fmt.Sprintf("Cannot authenticate user: %v", err.Error()))
	}
	if err != nil || valid == false {
		um.log.Info("User %v simian invalid", username)
		return user, valid
	}

	if guid != user.UserID {
		um.log.Error(fmt.Sprintf("Error: Authenticated user does not match local user"))
		return mgm.User{}, false
	}
	um.log.Info("User %v auth successful", username)
	return user, true
}

// SetPassword updates the credentials of the specified user
func (um Manager) SetPassword(userID uuid.UUID, password string) error {
	_, exists := um.GetUser(userID)
	if !exists {
		return errors.New("User not found")
	}
	return um.conn.SetPassword(userID, password)
}
