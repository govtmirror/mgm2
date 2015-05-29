package user

import (
	"errors"
	"time"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// Manager is an interface for working with pending and actual user accounts
type Manager interface {
	RequestControlPermission(mgm.Region, mgm.User) (mgm.Host, error)
	GetPendingUsers() ([]mgm.PendingUser, error)
	AddPendingUser(name string, email string, template string, password string, summary string) error
	GetEstates() ([]mgm.Estate, error)

	IsEmailUnique(string) (bool, error)
	IsNameUnique(name string) (bool, error)

	ValidatePasswordToken(userID uuid.UUID, token uuid.UUID) (bool, error)
	ScrubPasswordToken(token uuid.UUID) error
	CreatePasswordResetToken(userID uuid.UUID) (uuid.UUID, error)
}

// NewManager constructs a user.Manager for use
func NewManager(rMgr region.Manager, hMgr host.Manager, userConnector core.UserConnector, db database.Database, log core.Logger) Manager {
	um := userManager{}
	um.db = userDatabase{db}
	um.log = log
	um.conn = userConnector
	um.hMgr = hMgr
	um.rMgr = rMgr
	go um.expirePasswordTokens()
	return um
}

type userManager struct {
	rMgr region.Manager
	hMgr host.Manager
	db   userDatabase
	conn core.UserConnector
	log  core.Logger
}

func (um userManager) IsEmailUnique(email string) (bool, error) {
	//check for pending users in database
	unique, err := um.db.IsEmailUnique(email)
	if err != nil {
		return false, err
	}
	if !unique {
		return false, nil
	}
	//check email against simian accounts
	_, exists, err := um.conn.GetUserByEmail(email)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}
	return true, nil
}

func (um userManager) IsNameUnique(name string) (bool, error) {
	// check pending in database
	unique, err := um.db.IsNameUnique(name)
	if err != nil {
		return false, err
	}
	if !unique {
		return false, nil
	}
	//check with simian
	_, exists, err := um.conn.GetUserByName(name)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}
	return true, nil
}

func (um userManager) GetPendingUsers() ([]mgm.PendingUser, error) {
	return um.db.GetPendingUsers()
}

func (um userManager) AddPendingUser(name string, email string, template string, password string, summary string) error {
	return um.db.AddPendingUser(name, email, template, password, summary)
}

func (um userManager) ValidatePasswordToken(userID uuid.UUID, token uuid.UUID) (bool, error) {
	return um.db.ValidatePasswordToken(userID, token)
}

func (um userManager) ScrubPasswordToken(token uuid.UUID) error {
	return um.db.ScrubPasswordToken(token)
}

func (um userManager) CreatePasswordResetToken(userID uuid.UUID) (uuid.UUID, error) {
	return um.db.CreatePasswordResetToken(userID)
}

func (um userManager) GetEstates() ([]mgm.Estate, error) {
	return um.db.GetEstates()
}

func (um userManager) RequestControlPermission(region mgm.Region, user mgm.User) (mgm.Host, error) {
	h := mgm.Host{}

	//make sure user may control this region
	regions, err := um.rMgr.GetRegionsForUser(user.UserID)
	if err != nil {
		um.log.Error("Error retrieving regions for user: %v", err.Error())
		return h, err
	}

	for _, r := range regions {
		if r.UUID == region.UUID {
			h, err = um.hMgr.GetHostByID(r.Host)
			if err != nil {
				um.log.Error("Error host by address: %v", err.Error())
				return h, err
			}
			return h, nil
		}
	}
	return h, errors.New("Permission Denied")
}

func (um userManager) expirePasswordTokens() {
	for {
		um.db.ExpirePasswordTokens()
		time.Sleep(60 * time.Minute)
	}
}
