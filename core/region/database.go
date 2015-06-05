package region

import (
	"errors"

	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type regionDatabase struct {
	mysql database.Database
}

// GetRegionsForUser retrieves region records for a user where the user owns the estate they are in, or is a manager for said estate
func (db regionDatabase) GetRegionsForUser(guid uuid.UUID) ([]mgm.Region, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	var regions []mgm.Region

	rows, err := con.Query(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, host, EstateName from regions, estate_map, estate_settings " +
			"where estate_map.RegionID = regions.uuid AND estate_map.EstateID = estate_settings.EstateID AND uuid in " +
			"(SELECT RegionID FROM estate_map WHERE " +
			"EstateID in (SELECT EstateID FROM estate_settings WHERE EstateOwner=\"" + guid.String() + "\") OR " +
			"EstateID in (SELECT EstateID from estate_managers WHERE uuid=\"" + guid.String() + "\"))")
	defer rows.Close()
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return regions, nil
		}
		return nil, err
	}

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
			&r.EstateName,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}

// GetRegionByID retrieves a single region that matches the id given
func (db regionDatabase) GetRegionByID(id uuid.UUID) (mgm.Region, error) {
	r := mgm.Region{}
	con, err := db.mysql.GetConnection()
	if err != nil {
		return r, err
	}
	defer con.Close()

	err = con.QueryRow(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, host from regions "+
			"where uuid=?", id.String()).Scan(
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
		return r, err
	}
	if id != r.UUID {
		return r, errors.New("Region Not Found")
	}
	return r, nil
}

// GetRegions gets all region records from the database
func (db regionDatabase) GetRegions() ([]mgm.Region, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, host from regions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []mgm.Region
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
			rows.Close()
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}

func (db regionDatabase) GetDefaultConfigs() ([]mgm.ConfigOption, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	var cfgs []mgm.ConfigOption

	rows, err := con.Query("SELECT section, item, content FROM iniConfig WHERE region IS NULL")
	if err != nil {
		return nil, err
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
			return nil, err
		}
		cfgs = append(cfgs, c)
	}
	return cfgs, nil
}

func (db regionDatabase) GetConfigs(regionID uuid.UUID) ([]mgm.ConfigOption, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	var cfgs []mgm.ConfigOption

	rows, err := con.Query("SELECT section, item, content FROM iniConfig WHERE region=?", regionID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		c := mgm.ConfigOption{}
		err = rows.Scan(
			&c.Section,
			&c.Item,
			&c.Content,
		)
		c.Region = regionID
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, c)
	}
	return cfgs, nil
}

// GetRegionsOnHost retrieves all region records for a specified host
func (db regionDatabase) GetRegionsOnHost(host mgm.Host) ([]mgm.Region, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, host from regions "+
			"where host=?", host.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []mgm.Region
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
			rows.Close()
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}
