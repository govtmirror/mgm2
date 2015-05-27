package core

import (
	"errors"

	"github.com/M-O-S-E-S/mgm/mgm"
)

// RegionManager controls and notifies on region / estate changes and permissions
type RegionManager interface {
	RequestStart(mgm.Region, mgm.User) (mgm.Host, error)
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

func (rm regionMgr) RequestStart(region mgm.Region, user mgm.User) (mgm.Host, error) {
	host := mgm.Host{}

	return host, errors.New("Not Implemented")
}
