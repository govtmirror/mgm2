package persist

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
)

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
