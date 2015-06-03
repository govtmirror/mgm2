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
	rMgr.osdb = osdb
	rMgr.log = logger.Wrap("REGION", log)
	return rMgr
}

type regionMgr struct {
	simianURL string
	mgmURL    string
	db        regionDatabase
	osdb      database.Database
	log       logger.Log
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

	//map configs to eliminate duplicates, and so we can override values below
	configs := make(map[string]map[string]string)
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
	if _, ok := configs["Startup"]; !ok {
		configs["Startup"] = make(map[string]string)
	}
	configs["Startup"]["region_info_source"] = "filesystem"
	configs["Startup"]["Stats_URI"] = "jsonSimStats"
	if _, ok := configs["Network"]; !ok {
		configs["Network"] = make(map[string]string)
	}
	configs["Network"]["ConsoleUser"] = region.ConsoleUname.String()
	configs["Network"]["ConsolePass"] = region.ConsolePass.String()
	configs["Network"]["console_port"] = strconv.Itoa(region.ConsolePort)
	configs["Network"]["http_listener_port"] = strconv.Itoa(region.HTTPPort)
	configs["Network"]["ExternalHostNameForLSL"] = host.ExternalAddress
	if _, ok := configs["Messaging"]; !ok {
		configs["Messaging"] = make(map[string]string)
	}
	configs["Messaging"]["Gatekeeper"] = rm.simianURL
	configs["Messaging"]["OfflineMessageURL"] = rm.mgmURL + "messages"
	configs["Messaging"]["MuteListURL"] = rm.simianURL
	//if _, ok := configs["Groups"]; !ok {
	//	configs["Groups"] = make(map[string]string)
	//}
	//configs["Groups"]["GroupsServerURI"] = rm.simianURL
	//configs["Groups"]["XmlRpcServiceReadKey"] = groupsRead
	//configs["Groups"]["XmlRpcServiceWriteKey"] = groupsWrite
	if _, ok := configs["GridService"]; !ok {
		configs["GridService"] = make(map[string]string)
	}
	configs["GridService"]["GridServerURI"] = rm.simianURL
	configs["GridService"]["Gatekeeper"] = rm.simianURL
	if _, ok := configs["AssetService"]; !ok {
		configs["AssetService"] = make(map[string]string)
	}
	configs["AssetService"]["AssetServerURI"] = rm.simianURL
	if _, ok := configs["DatabaseService"]; !ok {
		configs["DatabaseService"] = make(map[string]string)
	}
	configs["DatabaseService"]["ConnectionString"] = rm.osdb.GetConnectionString()
	configs["DatabaseService"]["EstateConnectionString"] = rm.osdb.GetConnectionString()
	if _, ok := configs["InventoryService"]; !ok {
		configs["InventoryService"] = make(map[string]string)
	}
	configs["InventoryService"]["InventoryServerURI"] = rm.simianURL
	if _, ok := configs["GridInfo"]; !ok {
		configs["GridInfo"] = make(map[string]string)
	}
	configs["GridInfo"]["Gatekeeper"] = rm.simianURL
	if _, ok := configs["AvatarService"]; !ok {
		configs["AvatarService"] = make(map[string]string)
	}
	configs["AvatarService"]["AvatarServerURI"] = rm.simianURL
	if _, ok := configs["PresenceService"]; !ok {
		configs["PresenceService"] = make(map[string]string)
	}
	configs["PresenceService"]["PresenceServerURI"] = rm.simianURL
	if _, ok := configs["UserAccountService"]; !ok {
		configs["UserAccountService"] = make(map[string]string)
	}
	configs["UserAccountService"]["UserAccountServerURI"] = rm.simianURL
	if _, ok := configs["GridUserService"]; !ok {
		configs["GridUserService"] = make(map[string]string)
	}
	configs["GridUserService"]["GridUserServerURI"] = rm.simianURL
	if _, ok := configs["AuthenticationService"]; !ok {
		configs["AuthenticationService"] = make(map[string]string)
	}
	configs["AuthenticationService"]["AuthenticationServerURI"] = rm.simianURL
	if _, ok := configs["FriendsService"]; !ok {
		configs["FriendsService"] = make(map[string]string)
	}
	configs["FriendsService"]["FriendsServerURI"] = rm.simianURL
	//if _, ok := configs["HGInventoryAccessModule"]; !ok {
	//	configs["HGInventoryAccessModule"] = make(map[string]string)
	//}
	//configs["HGInventoryAccessModule"]["HomeURI"] = rm.simianURL
	//configs["HGInventoryAccessModule"]["GateKeeper"] = rm.simianURL
	//if _, ok := configs["HGAssetService"]; !ok {
	//	configs["HGAssetService"] = make(map[string]string)
	//}
	//configs["HGAssetService"]["HomeURI"] = rm.simianURL
	if _, ok := configs["UserAgentService"]; !ok {
		configs["UserAgentService"] = make(map[string]string)
	}
	configs["UserAgentService"]["UserAgentServiceURI"] = rm.simianURL
	if _, ok := configs["MapImageService"]; !ok {
		configs["MapImageService"] = make(map[string]string)
	}
	configs["MapImageService"]["MapImageServiceURI"] = rm.simianURL
	if _, ok := configs["SimianGrid"]; !ok {
		configs["SimianGrid"] = make(map[string]string)
	}
	configs["SimianGrid"]["SimianServiceURL"] = rm.simianURL
	configs["SimianGrid"]["SimulatorCapability"] = "00000000-0000-0000-0000-000000000000"
	if _, ok := configs["SimianGridMaptiles"]; !ok {
		configs["SimianGridMaptiles"] = make(map[string]string)
	}
	configs["SimianGridMaptiles"]["Enabled"] = "true"
	configs["SimianGridMaptiles"]["MaptileURL"] = rm.simianURL
	configs["SimianGridMaptiles"]["RefreshTime"] = "7200"
	if _, ok := configs["Terrain"]; !ok {
		configs["Terrain"] = make(map[string]string)
	}
	configs["Terrain"]["SendTerrainUpdatesByViewDistance"] = "true"

	for section, m := range configs {
		for item, content := range m {
			result = append(result, mgm.ConfigOption{region.UUID, section, item, content})
		}
	}

	return result, nil
}
