package user

import (
	"fmt"
	"sync"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/core/job"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/m-o-s-e-s/mgm/sql"
	"github.com/satori/go.uuid"
)

type notifier interface {
}

// NewManager constructs a user.Manager for use
func NewManager(rMgr *region.Manager, hMgr *host.Manager, jMgr *job.Manager, userConnector core.UserConnector, pers *sql.MGMDB, notify notifier, log logger.Log) *Manager {
	um := Manager{}
	um.log = logger.Wrap("USER", log)
	um.conn = userConnector
	um.hMgr = hMgr
	um.rMgr = rMgr
	um.jMgr = jMgr
	um.mgm = pers
	um.users = make(map[uuid.UUID]mgm.User)
	um.uMutex = &sync.Mutex{}
	um.pendingUsers = make(map[string]mgm.PendingUser)
	um.puMutex = &sync.Mutex{}
	um.estates = make(map[int64]mgm.Estate)
	um.estateMutex = &sync.Mutex{}
	um.groups = make(map[uuid.UUID]mgm.Group)
	um.groupMutex = &sync.Mutex{}
	um.notify = notify

	// Get users from simian
	users, err := userConnector.GetUsers()
	if err != nil {
		um.log.Fatal("Cannot get users from simian: ", err.Error())
	}
	for _, u := range users {
		um.users[u.UserID] = u
	}

	// get pending users from mysql
	for _, u := range pers.QueryPendingUsers() {
		um.pendingUsers[u.Email] = u
	}

	//get estates from mysql
	for _, e := range pers.QueryEstates() {
		um.estates[e.ID] = e
	}

	// Get groups from simian
	groups, err := userConnector.GetGroups()
	if err != nil {
		um.log.Fatal("Cannot get groups from simian: ", err.Error())
	}
	for _, g := range groups {
		um.groups[g.ID] = g
	}

	return &um
}

// Manager is a central access point for user functions
type Manager struct {
	rMgr         *region.Manager
	mgm          *sql.MGMDB
	hMgr         *host.Manager
	jMgr         *job.Manager
	notify       notifier
	conn         core.UserConnector
	log          logger.Log
	pendingUsers map[string]mgm.PendingUser
	puMutex      *sync.Mutex
	users        map[uuid.UUID]mgm.User
	uMutex       *sync.Mutex
	estates      map[int64]mgm.Estate
	estateMutex  *sync.Mutex
	groups       map[uuid.UUID]mgm.Group
	groupMutex   *sync.Mutex
}

// TestAdminAccess ensures that at least one admin account exists, creating a default if none is found
func (m *Manager) TestAdminAccess() error {
	m.uMutex.Lock()
	defer m.uMutex.Unlock()
	adminCount := 0
	for _, u := range m.users {
		if u.AccessLevel >= 250 {
			adminCount++
		}
	}
	if adminCount > 0 {
		m.log.Info("Administrative users present")
		return nil
	}
	name := "mgm admin"
	email := "~"
	m.log.Info("No administrative access found, creating default admin account")
	uuid, err := m.conn.CreateUserEntry(name, email)
	if err != nil {
		return err
	}
	success, err := m.conn.CreateUserInventory(uuid, "default")
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("Cannot create admin user inventory")
	}
	err = m.conn.SetPassword(uuid, "password")
	if err != nil {
		return err
	}
	err = m.conn.UpdateUser(name, email, uuid, 250)
	return err
}
