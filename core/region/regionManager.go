package region

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/m-o-s-e-s/mgm/sql"
	"github.com/satori/go.uuid"
)

type notifier interface {
}

// NewManager constructs a RegionManager for use
func NewManager(mgmURL string, simianURL string, pers *sql.MGMDB, osdb sql.Database, notify notifier, log logger.Log) *Manager {
	rMgr := Manager{}
	rMgr.simianURL = simianURL
	rMgr.mgmURL = mgmURL
	rMgr.mgm = pers
	rMgr.osdb = osdb
	rMgr.log = logger.Wrap("REGION", log)
	rMgr.regions = make(map[uuid.UUID]mgm.Region)
	rMgr.regionStats = make(map[uuid.UUID]mgm.RegionStat)
	rMgr.rMutex = &sync.Mutex{}
	rMgr.rsMutex = &sync.Mutex{}
	rMgr.notify = notify

	for _, r := range pers.QueryRegions() {
		rMgr.regions[r.UUID] = r
		rMgr.regionStats[r.UUID] = mgm.RegionStat{}
	}

	return &rMgr
}

// Manager is a central access point for Region actions
type Manager struct {
	simianURL   string
	mgmURL      string
	osdb        sql.Database
	mgm         *sql.MGMDB
	notify      notifier
	log         logger.Log
	regions     map[uuid.UUID]mgm.Region
	rMutex      *sync.Mutex
	regionStats map[uuid.UUID]mgm.RegionStat
	rsMutex     *sync.Mutex
}

// GetRegions get a slice of all regions from cache
func (m Manager) GetRegions() []mgm.Region {
	m.rMutex.Lock()
	defer m.rMutex.Unlock()
	t := []mgm.Region{}
	for _, r := range m.regions {
		t = append(t, r)
	}
	return t
}

// GetRegionStats get a slice of all region stats from cache
func (m Manager) GetRegionStats() []mgm.RegionStat {
	m.rsMutex.Lock()
	defer m.rsMutex.Unlock()
	t := []mgm.RegionStat{}
	for _, r := range m.regionStats {
		t = append(t, r)
	}
	return t
}

// GetDefaultConfigs retrieves the default region configuration
func (m Manager) GetDefaultConfigs() []mgm.ConfigOption {
	return m.mgm.QueryDefaultConfigs()
}

// GetConfigs retrieves region-specific configuration
func (m Manager) GetConfigs(id uuid.UUID) []mgm.ConfigOption {
	return m.mgm.QueryConfigs(id)
}

// ServeConfigs generates a list of configuration options to feed to a region before it starts
func (m Manager) ServeConfigs(region mgm.Region, host mgm.Host) []mgm.ConfigOption {
	var result []mgm.ConfigOption

	gridURL := fmt.Sprintf("http://%v/Grid/", m.simianURL)

	defaultConfigs := m.mgm.QueryDefaultConfigs()
	regionConfigs := m.mgm.QueryConfigs(region.UUID)

	configs := make(map[string]map[string]string)

	// initialize sections, so we dont wipe them out when we force values below
	//configs["Const"] = make(map[string]string)
	configs["Startup"] = make(map[string]string)
	configs["Permissions"] = make(map[string]string)
	configs["Network"] = make(map[string]string)
	configs["ClientStack.LindenCaps"] = make(map[string]string)
	configs["Messaging"] = make(map[string]string)
	configs["Groups"] = make(map[string]string)
	configs["Terrain"] = make(map[string]string)
	configs["Architecture"] = make(map[string]string)
	configs["DatabaseService"] = make(map[string]string)
	configs["Modules"] = make(map[string]string)
	configs["AssetService"] = make(map[string]string)
	configs["InventoryService"] = make(map[string]string)
	configs["GridInfo"] = make(map[string]string)
	configs["GridService"] = make(map[string]string)
	configs["AvatarService"] = make(map[string]string)
	configs["PresenceService"] = make(map[string]string)
	configs["UserAccountService"] = make(map[string]string)
	configs["AuthenticationService"] = make(map[string]string)
	configs["FriendsService"] = make(map[string]string)
	configs["MapImageService"] = make(map[string]string)
	configs["OSSL"] = make(map[string]string)
	configs["SimianGrid"] = make(map[string]string)
	configs["GridUserService"] = make(map[string]string)

	//map configs to eliminate duplicates
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
	//configs["Const"]["SimianURL"] = "http://" + rm.simianURL + "/Grid/"
	//configs["Const"]["MGMURL"] = "http://" + rm.mgmURL
	/*
			[Const]
		    ;# {BaseURL} {} {BaseURL} {"http://example.com","http://127.0.0.1"} "http://127.0.0.1"
		    BaseURL = http://127.0.0.1

		    ;# {PublicPort} {} {PublicPort} {8002} "8002"
		    PublicPort = "8002"

		    ;# {PrivatePort} {} {PrivatePort} {8003} "8003"
		    PrivatePort = "8003"
	*/

	configs["Startup"]["Stats_URI"] = "jsonSimStats"

	configs["Permissions"]["allow_grid_gods"] = "true"

	configs["Network"]["ConsoleUser"] = region.ConsoleUname.String()
	configs["Network"]["ConsolePass"] = region.ConsolePass.String()
	configs["Network"]["console_port"] = strconv.Itoa(region.ConsolePort)
	configs["Network"]["http_listener_port"] = strconv.Itoa(region.HTTPPort)
	configs["Network"]["ExternalHostNameForLSL"] = host.ExternalAddress
	configs["Network"]["shard"] = "OpenSim"

	configs["ClientStack.LindenCaps"]["Cap_GetTexture"] = fmt.Sprintf("http://%v/GridPublic/GetTexture/", m.simianURL)
	configs["ClientStack.LindenCaps"]["Cap_GetMesh"] = fmt.Sprintf("http://%v/GridPublic/GetMesh", m.simianURL)
	configs["ClientStack.LindenCaps"]["Cap_AvatarPickerSearch"] = "localhost"
	configs["ClientStack.LindenCaps"]["Cap_GetDisplayNames"] = "localhost"

	/*
						[Messaging]
						  ; OfflineMessageModule = OfflineMessageModule
						  ; OfflineMessageModule = "Offline Message Module V2"

						  ; OfflineMessageURL = ${Const|BaseURL}/Offline.php
					    ; OfflineMessageURL = ${Const|BaseURL}:${Const|PrivatePort}

						  ; StorageProvider = OpenSim.Data.MySQL.dll
						  ; MuteListModule = MuteListModule
						  ; MuteListURL = http://yourserver/Mute.php
						  ; ForwardOfflineGroupMessages = true

						[FreeSwitchVoice]
			    		; Enabled = false
			    		; LocalServiceModule = OpenSim.Services.Connectors.dll:RemoteFreeswitchConnector
		    			; FreeswitchServiceURL = http://my.grid.server:8004/fsapi

	*/

	configs["Groups"]["LevelGroupCreate"] = "0"
	configs["Groups"]["Module"] = "GroupsModule"
	configs["Groups"]["StorageProvider"] = "OpenSim.Data.MySQL.dll"
	configs["Groups"]["ServicesConnectorModule"] = "SimianGroupsServiceConnector"
	configs["Groups"]["GroupsServerURI"] = gridURL
	configs["Groups"]["MessagingModule"] = "GroupsMessagingModule"

	configs["Terrain"]["InitialTerrain"] = "flat"

	/*
			[UserProfiles]
		  ;; ProfileServiceURL = ${Const|BaseURL}:${Const|PublicPort}
	*/

	configs["Architecture"]["Include-Architecture"] = "config-include/SimianGrid.ini"

	configs["DatabaseService"]["StorageProvider"] = "OpenSim.Data.MySQL.dll"
	configs["DatabaseService"]["ConnectionString"] = m.osdb.GetConnectionString()

	configs["Modules"]["AssetCaching"] = "FlotsamAssetCache"
	configs["Modules"]["Include-FlotsamCache"] = "config-include/FlotsamCache.ini"

	configs["AssetService"]["DefaultAssetLoader"] = "OpenSim.Framework.AssetLoader.Filesystem.dll"
	configs["AssetService"]["AssetLoaderArgs"] = "assets/AssetSets.xml"
	configs["AssetService"]["AssetServerURI"] = gridURL

	configs["InventoryService"]["InventoryServerURI"] = gridURL

	configs["GridInfo"]["GridInfoURI"] = m.mgmURL

	configs["GridService"]["GridServerURI"] = gridURL

	//configs["EstateDataStore"]["LocalServiceModule"] = "OpenSim.Services.Connectors.dll:EstateDataRemoteConnector"
	//configs["EstateService"]["EstateServerURI"] = "${Const|BaseURL}:${Const|PrivatePort}"

	configs["AvatarService"]["AvatarServerURI"] = gridURL

	configs["PresenceService"]["PresenceServerURI"] = gridURL

	configs["UserAccountService"]["UserAccountServerURI"] = gridURL

	configs["GridUserService"]["GridUserServerURI"] = gridURL

	configs["AuthenticationService"]["AuthenticationServerURI"] = gridURL

	configs["FriendsService"]["FriendsServerURI"] = gridURL

	configs["MapImageService"]["MapImageServerURI"] = gridURL

	configs["OSSL"]["Include-osslEnable"] = "config-include/osslEnable.ini"

	configs["SimianGrid"]["SimianServiceURL"] = gridURL

	//convert map into a single slice of ConfigOption
	for section, m := range configs {
		for item, content := range m {
			result = append(result,
				mgm.ConfigOption{
					Region:  region.UUID,
					Section: section,
					Item:    item,
					Content: content,
				},
			)
		}
	}

	return result
}
