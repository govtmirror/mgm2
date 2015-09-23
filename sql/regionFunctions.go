package sql

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// PersistRegion commits a region record to the database
func (m MGMDB) PersistRegion(region mgm.Region) {
	con, err := m.db.getConnection()
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

// QueryRegions reads all region records from the database
func (m MGMDB) QueryRegions() []mgm.Region {
	var regions []mgm.Region
	con, err := m.db.getConnection()
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

// QueryDefaultConfigs reads the current default configs from the database
func (m MGMDB) QueryDefaultConfigs() []mgm.ConfigOption {
	cfgs := []mgm.ConfigOption{}
	con, err := m.db.getConnection()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to database: %v", err.Error()))
		return cfgs
	}
	defer con.Close()

	rows, err := con.Query("SELECT section, item, content FROM iniConfig WHERE region IS NULL")
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting default configs: %v", err.Error()))
		return cfgs
	}
	defer rows.Close()

	for rows.Next() {
		c := mgm.ConfigOption{}
		err = rows.Scan(
			&c.Section,
			&c.Item,
			&c.Content,
		)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error parsing default configs: %v", err.Error()))
			return cfgs
		}
		cfgs = append(cfgs, c)
	}
	return cfgs
}

// QueryConfigs reads the current specific configs for a given region
func (m MGMDB) QueryConfigs(r uuid.UUID) []mgm.ConfigOption {
	cfgs := []mgm.ConfigOption{}
	con, err := m.db.getConnection()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to database: %v", err.Error()))
		return cfgs
	}
	defer con.Close()

	rows, err := con.Query("SELECT section, item, content FROM iniConfig WHERE region=?", r.String())
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting configs for %v: %v", r.String(), err.Error()))
		return cfgs
	}
	defer rows.Close()

	for rows.Next() {
		c := mgm.ConfigOption{}
		err = rows.Scan(
			&c.Section,
			&c.Item,
			&c.Content,
		)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error parsing configs for %v: %v", r.String(), err.Error()))
			return cfgs
		}
		cfgs = append(cfgs, c)
	}
	return cfgs
}
