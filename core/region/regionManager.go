package region

import (
	"errors"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/mgm"
)

// Manager controls and notifies on region / estate changes and permissions
type Manager interface {
	RequestControlPermission(mgm.Region, mgm.User) (mgm.Host, error)
}

// NewManager constructs a RegionManager for use
func NewManager(nMgr host.Manager, db core.Database, log core.Logger) Manager {
	rMgr := regionMgr{}
	rMgr.nodeMgr = nMgr
	rMgr.db = db
	rMgr.log = log
	return rMgr
}

type regionMgr struct {
	nodeMgr host.Manager
	db      core.Database
	log     core.Logger
}

func (rm regionMgr) RequestControlPermission(region mgm.Region, user mgm.User) (mgm.Host, error) {
	h := mgm.Host{}

	//make sure user may control this region
	regions, err := rm.db.GetRegionsForUser(user.UserID)
	if err != nil {
		rm.log.Error("Error retrieving regions for user: %v", err.Error())
		return h, err
	}

	for _, r := range regions {
		if r.UUID == region.UUID {
			h, err = rm.db.GetHostByAddress(r.SlaveAddress)
			if err != nil {
				rm.log.Error("Error host by address: %v", err.Error())
				return h, err
			}
			return h, nil
		}
	}
	return h, errors.New("Permission Denied")
}
