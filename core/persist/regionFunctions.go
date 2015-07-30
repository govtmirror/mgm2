package persist

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

func (m mgmDB) persistRegion(region mgm.Region) {
	con, err := m.db.GetConnection()
	if err != nil {
		errMsg := fmt.Sprintf("Error connecting to database: %v", err.Error())
		log.Fatal(errMsg)
	}
	defer con.Close()

	errMsg := fmt.Sprintf("Persisting region %v", region.UUID)
	m.log.Info(errMsg)

	_, err = con.Exec("REPLACE INTO regions VALUES (?,?,?,?,?,?,?,?,?,?)",
		region.UUID.String(),
		region.Name,
		region.Size,
		region.HTTPPort,
		region.ConsolePort,
		region.ConsoleUname.String(),
		region.ConsolePass.String(),
		region.LocX,
		region.LocY,
		region.Host)
	if err != nil {
		errMsg := fmt.Sprintf("Error updating region: %v", err.Error())
		m.log.Error(errMsg)
	}
}

func (m mgmDB) queryRegions() []mgm.Region {
	var regions []mgm.Region
	con, err := m.db.GetConnection()
	if err != nil {
		errMsg := fmt.Sprintf("Error connecting to database: %v", err.Error())
		log.Fatal(errMsg)
		return regions
	}
	defer con.Close()
	rows, err := con.Query(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, host from regions")
	if err != nil {
		errMsg := fmt.Sprintf("Error reading regions: %v", err.Error())
		m.log.Error(errMsg)
		return regions
	}
	defer rows.Close()
	for rows.Next() {
		r := mgm.Region{}
		err = rows.Scan(
			&r.UUID,
			&r.Name,
			&r.Size,
			&r.HTTPPort,
			&r.ConsolePort,
			&r.ConsoleUname,
			&r.ConsolePass,
			&r.LocX,
			&r.LocY,
			&r.Host,
		)
		if err != nil {
			errMsg := fmt.Sprintf("Error scanning regions: %v", err.Error())
			m.log.Error(errMsg)
			return regions
		}
		regions = append(regions, r)
	}

	return regions
}

func (m mgmDB) GetDefaultConfigs() []mgm.ConfigOption {
	var configs []mgm.ConfigOption
	r := mgmReq{}
	r.request = "GetDefaultConfigs"
	r.result = make(chan interface{}, 64)
	m.reqs <- r
	for {
		h, ok := <-r.result
		if !ok {
			return configs
		}
		configs = append(configs, h.(mgm.ConfigOption))
	}
}

func (m mgmDB) GetConfigs(region mgm.Region) []mgm.ConfigOption {
	var configs []mgm.ConfigOption
	r := mgmReq{}
	r.request = "GetConfigs"
	r.object = region
	r.result = make(chan interface{}, 64)
	m.reqs <- r
	for {
		h, ok := <-r.result
		if !ok {
			return configs
		}
		configs = append(configs, h.(mgm.ConfigOption))
	}
}

func (m mgmDB) GetRegions() []mgm.Region {
	var regions []mgm.Region
	r := mgmReq{}
	r.request = "GetRegions"
	r.result = make(chan interface{}, 64)
	m.reqs <- r
	for {
		h, ok := <-r.result
		if !ok {
			return regions
		}
		regions = append(regions, h.(mgm.Region))
	}
}

func (m mgmDB) GetRegion(id uuid.UUID) (mgm.Region, bool) {
	for _, r := range m.GetRegions() {
		if r.UUID == id {
			return r, true
		}
	}
	return mgm.Region{}, false
}

func (m mgmDB) GetRegionStats() []mgm.RegionStat {
	var stats []mgm.RegionStat
	r := mgmReq{}
	r.request = "GetRegionStats"
	r.result = make(chan interface{}, 64)
	m.reqs <- r
	for {
		h, ok := <-r.result
		if !ok {
			return stats
		}
		stats = append(stats, h.(mgm.RegionStat))
	}
}

func (m mgmDB) GetRegionStat(id uuid.UUID) (mgm.RegionStat, bool) {
	for _, st := range m.GetRegionStats() {
		if st.UUID == id {
			return st, true
		}
	}
	return mgm.RegionStat{}, false
}

func (m mgmDB) UpdateRegion(region mgm.Region) {
	r := mgmReq{}
	r.request = "UpdateRegion"
	r.object = region
	m.reqs <- r
}

func (m mgmDB) UpdateRegionStat(stat mgm.RegionStat) {
	r := mgmReq{}
	r.request = "UpdateRegionStat"
	r.object = stat
	m.reqs <- r
}

func (m mgmDB) RemoveRegion(region mgm.Region) {
	r := mgmReq{}
	r.request = "RemoveRegion"
	r.object = region
	m.reqs <- r
}

func (m mgmDB) MoveRegionToHost(r mgm.Region, h mgm.Host) {
	req := mgmReq{}
	req.request = "MoveRegionToHost"
	req.object = r
	req.target = h
	req.result = make(chan interface{}, 4)
	m.reqs <- req
}

func (m mgmDB) MoveRegionToEstate(r mgm.Region, e mgm.Estate) {
	req := mgmReq{}
	req.request = "MoveRegionToEstate"
	req.object = r
	req.target = e
	req.result = make(chan interface{}, 4)
	m.reqs <- req
}
