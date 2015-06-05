package region

import (
	"strconv"

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
	ServeConfigs(mgm.Region, mgm.Host) ([]mgm.ConfigOption, error)
}

// NewManager constructs a RegionManager for use
func NewManager(mgmURL string, simianURL string, db database.Database, osdb database.Database, log logger.Log) Manager {
	rMgr := regionMgr{}
	rMgr.simianURL = simianURL
	rMgr.mgmURL = mgmURL
	rMgr.db = regionDatabase{db}
	rMgr.osdb = simDatabase{osdb}
	rMgr.log = logger.Wrap("REGION", log)
	return rMgr
}

type regionMgr struct {
	simianURL string
	mgmURL    string
	db        regionDatabase
	osdb      simDatabase
	log       logger.Log
}

func (rm regionMgr) GetRegionsForUser(guid uuid.UUID) ([]mgm.Region, error) {
	rgs, err := rm.db.GetRegionsForUser(guid)
	if err != nil {
		return nil, err
	}
	for i, r := range rgs {
		n, err := rm.osdb.GetEstateNameForRegion(r)
		if err != nil {
			rm.log.Error("Error getting estate for region: %s", err.Error())
		} else {
			rgs[i].EstateName = n
		}
	}
	return rgs, nil
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
	rgs, err := rm.db.GetRegions()
	if err != nil {
		return nil, err
	}
	for i, r := range rgs {
		n, err := rm.osdb.GetEstateNameForRegion(r)
		if err != nil {
			rm.log.Error("Error getting estate for region: %s", err.Error())
		} else {
			rgs[i].EstateName = n
		}
	}
	return rgs, nil
}

func (rm regionMgr) GetRegionsOnHost(host mgm.Host) ([]mgm.Region, error) {
	rgs, err := rm.db.GetRegionsOnHost(host)
	if err != nil {
		return nil, err
	}
	for i, r := range rgs {
		n, err := rm.osdb.GetEstateNameForRegion(r)
		if err != nil {
			rm.log.Error("Error getting estate for region: %s", err.Error())
		} else {
			rgs[i].EstateName = n
		}
	}
	return rgs, nil
}

func (rm regionMgr) ServeConfigs(region mgm.Region, host mgm.Host) ([]mgm.ConfigOption, error) {
	var result []mgm.ConfigOption

	defaultConfigs, err := rm.GetDefaultConfigs()
	if err != nil {
		return result, err
	}
	regionConfigs, err := rm.GetConfigs(region.UUID)
	if err != nil {
		return result, err
	}

	configs := make(map[string]map[string]string)

	//insert initial values that may be overridden
	configs["Const"] = make(map[string]string)
	configs["Startup"] = make(map[string]string)
	configs["Network"] = make(map[string]string)
	configs["ClientStack.LindenCaps"] = make(map[string]string)
	configs["DatabaseService"] = make(map[string]string)

	//map configs to eliminate duplicates, and so we can override values below
	for _, cfg := range defaultConfigs {
		if _, ok := configs[cfg.Section]; !ok {
			configs[cfg.Section] = make(map[string]string)
		}
		configs[cfg.Section][cfg.Item] = cfg.Content
	}
	for _, cfg := range regionConfigs {
		if _, ok := configs[cfg.Section]; !ok {
			configs[cfg.Section] = make(map[string]string)
		}
		configs[cfg.Section][cfg.Item] = cfg.Content
	}

	//override fields with installation-static options
	configs["Const"]["SimianURL"] = "http://" + rm.simianURL + "/Grid/"
	configs["Const"]["MGMURL"] = "http://" + rm.mgmURL

	configs["Startup"]["PIDFile"] = "moses.pid"
	configs["Startup"]["region_info_source"] = "filesystem"
	configs["Startup"]["allow_regionless"] = "false"
	configs["Startup"]["Stats_URI"] = "jsonSimStats"
	configs["Startup"]["OutboundDisallowForUserScripts"] = "0.0.0.0/8|10.0.0.0/8|100.64.0.0/10|127.0.0.0/8|169.254.0.0/16|172.16.0.0/12|192.0.0.0/24|192.0.2.0/24|192.88.99.0/24|192.168.0.0/16|198.18.0.0/15|198.51.100.0/24|203.0.113.0/24|224.0.0.0/4|240.0.0.0/4|255.255.255.255/32"

	configs["Network"]["ConsoleUser"] = region.ConsoleUname.String()
	configs["Network"]["ConsolePass"] = region.ConsolePass.String()
	configs["Network"]["console_port"] = strconv.Itoa(region.ConsolePort)
	configs["Network"]["http_listener_port"] = strconv.Itoa(region.HTTPPort)
	configs["Network"]["ExternalHostNameForLSL"] = host.ExternalAddress

	configs["ClientStack.LindenCaps"]["Cap_CopyInventoryFromNotecard"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_EnvironmentSettings"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_EventQueueGet"] = "localhost"
	configs["ClientStack.LindenCaps"]["ObjectMedia"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_ObjectMediaNavigate"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_GetDisplayNames"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_GetTexture"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_GetMesh"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_MapLayer"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_MapLayerGod"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_NewFileAgentInventory"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_NewFileAgentInventoryVariablePrice"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_ObjectAdd"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_ParcelPropertiesUpdate"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_RemoteParcelRequest"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_UpdateNotecardAgentInventory"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_UpdateScriptAgent"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_UpdateNotecardTaskInventory"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_UpdateScriptTask"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_UploadBakedTexture"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_UploadObjectAsset"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_AvatarPickerSearch"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_FetchInventoryDescendents2"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_FetchInventory2"] = "localhost"

	configs["DatabaseService"]["StorageProvider"] = "OpenSim.Data.MySQL.dll"
	configs["DatabaseService"]["ConnectionString"] = rm.osdb.GetConnectionString()

	//convert map into a single slice of ConfigOption
	for section, m := range configs {
		for item, content := range m {
			result = append(result, mgm.ConfigOption{region.UUID, section, item, content})
		}
	}

	return result, nil
}
