package core

import (
	"errors"

	"github.com/M-O-S-E-S/mgm/mgm"
)

// RegionManager controls and notifies on region / estate changes and permissions
type RegionManager interface {
	RequestControlPermission(mgm.Region, mgm.User) (mgm.Host, error)
}

// NewRegionManager constructs a RegionManager for use
func NewRegionManager(nMgr NodeManager, db Database, log Logger) RegionManager {
	rMgr := regionMgr{}
	rMgr.nodeMgr = nMgr
	rMgr.db = db
	rMgr.log = log
	return rMgr
}

type regionMgr struct {
	nodeMgr NodeManager
	db      Database
	log     Logger
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
