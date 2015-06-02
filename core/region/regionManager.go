package region

import (
	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// Manager controls and notifies on region / estate changes and permissions
type Manager interface {
	GetRegionsForUser(guid uuid.UUID) ([]mgm.Region, error)
	GetRegionByID(id uuid.UUID) (mgm.Region, error)
	GetDefaultConfigs() ([]mgm.ConfigOption, error)
	GetConfigs(regionID uuid.UUID) ([]mgm.ConfigOption, error)
	GetRegions() ([]mgm.Region, error)
	GetRegionsOnHost(host mgm.Host) ([]mgm.Region, error)
}

// NewManager constructs a RegionManager for use
func NewManager(db database.Database, log logger.Log) Manager {
	rMgr := regionMgr{}
	rMgr.db = regionDatabase{db}
	rMgr.log = logger.Wrap("REGION", log)
	return rMgr
}

type regionMgr struct {
	db  regionDatabase
	log logger.Log
}

func (rm regionMgr) GetRegionsForUser(guid uuid.UUID) ([]mgm.Region, error) {
	return rm.db.GetRegionsForUser(guid)
}

func (rm regionMgr) GetRegionByID(id uuid.UUID) (mgm.Region, error) {
	return rm.db.GetRegionByID(id)
}

func (rm regionMgr) GetDefaultConfigs() ([]mgm.ConfigOption, error) {
	return rm.db.GetDefaultConfigs()
}

func (rm regionMgr) GetConfigs(regionID uuid.UUID) ([]mgm.ConfigOption, error) {
	return rm.db.GetConfigs(regionID)
}

func (rm regionMgr) GetRegions() ([]mgm.Region, error) {
	return rm.db.GetRegions()
}

func (rm regionMgr) GetRegionsOnHost(host mgm.Host) ([]mgm.Region, error) {
	return rm.db.GetRegionsOnHost(host)
}
