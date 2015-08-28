package user

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// GetUsers gets an array of all current users
func (m Manager) GetUsers() []mgm.User {
	m.uMutex.Lock()
	defer m.uMutex.Unlock()
	t := []mgm.User{}
	for _, user := range m.users {
		t = append(t, user)
	}
	return t
}

// GetUser gets a user matching a specific uuid
func (m Manager) GetUser(id uuid.UUID) (mgm.User, bool) {
	m.uMutex.Lock()
	defer m.uMutex.Unlock()
	u, ok := m.users[id]
	return u, ok
}

// UserIsAdmin is a utility function to get the current admin status of the specified user
func (m Manager) UserIsAdmin(id uuid.UUID) bool {
	m.uMutex.Lock()
	defer m.uMutex.Unlock()
	u, ok := m.users[id]
	if !ok || u.AccessLevel < 250 {
		return false
	}
	return true
}

// GetPendingUsers gets a list of all pending users
func (m Manager) GetPendingUsers() []mgm.PendingUser {
	m.puMutex.Lock()
	defer m.puMutex.Unlock()
	t := []mgm.PendingUser{}
	for _, user := range m.pendingUsers {
		t = append(t, user)
	}
	return t
}

// AddPendingUser inserts a specified pending user into the cache
func (m Manager) AddPendingUser(name string, email string, template string, password string, summary string) {
	hasher := md5.New()
	hasher.Write([]byte(password))
	creds := hex.EncodeToString(hasher.Sum(nil))

	pu := mgm.PendingUser{}
	pu.Name = name
	pu.Email = email
	pu.Gender = template
	pu.PasswordHash = creds
	pu.Summary = summary

	m.puMutex.Lock()
	defer m.puMutex.Unlock()
	m.pendingUsers[pu.Email] = pu
}

// UpdateUser caches and persists a modified user record
func (m Manager) UpdateUser(user mgm.User) {
	m.uMutex.Lock()
	defer m.uMutex.Unlock()
	m.users[user.UserID] = user
	//not implemented m.mgm.UpdateUser(user)
}
