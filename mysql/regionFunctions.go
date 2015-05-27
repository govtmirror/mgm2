package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/M-O-S-E-S/mgm/mgm"
	//import mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
)

// GetRegions gets all region records from the database
func (db db) GetRegions() ([]mgm.Region, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning, EstateName from regions, estate_map, estate_settings " +
			"where estate_map.RegionID = regions.uuid AND estate_map.EstateID = estate_settings.EstateID")
	defer rows.Close()
	if err != nil {
		db.log.Error("Error in database query: ", err.Error())
		return nil, err
	}

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
			&r.ExternalAddress,
			&r.SlaveAddress,
			&r.IsRunning,
			&r.EstateName,
		)
		if err != nil {
			rows.Close()
			db.log.Error("Error in database query: ", err.Error())
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}

// GetRegionByID retrieves a single region that matches the id given
func (db db) GetRegionByID(id uuid.UUID) (mgm.Region, error) {
	r := mgm.Region{}
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return r, err
	}
	defer con.Close()

	err = con.QueryRow(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning from regions "+
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
		&r.ExternalAddress,
		&r.SlaveAddress,
		&r.IsRunning,
	)
	if err != nil {
		db.log.Error("Error in database query: ", err.Error())
		return r, err
	}
	if id != r.UUID {
		return r, errors.New("Region Not Found")
	}
	return r, nil
}

// GetRegionsOnHost retrieves all region records for a specified host
func (db db) GetRegionsOnHost(host mgm.Host) ([]mgm.Region, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning from regions "+
			"where slaveAddress=?", host.Address)
	defer rows.Close()
	if err != nil {
		db.log.Error("Error in database query: ", err.Error())
		return nil, err
	}

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
			&r.ExternalAddress,
			&r.SlaveAddress,
			&r.IsRunning,
		)
		if err != nil {
			rows.Close()
			db.log.Error("Error in database query: ", err.Error())
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}

// GetRegionsForUser retrieves region records for a user where the user owns the estate they are in, or is a manager for said estate
func (db db) GetRegionsForUser(guid uuid.UUID) ([]mgm.Region, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning, EstateName from regions, estate_map, estate_settings " +
			"where estate_map.RegionID = regions.uuid AND estate_map.EstateID = estate_settings.EstateID AND uuid in " +
			"(SELECT RegionID FROM estate_map WHERE " +
			"EstateID in (SELECT EstateID FROM estate_settings WHERE EstateOwner=\"" + guid.String() + "\") OR " +
			"EstateID in (SELECT EstateID from estate_managers WHERE uuid=\"" + guid.String() + "\"))")
	defer rows.Close()
	if err != nil {
		db.log.Error("Error in database query: ", err.Error())
		return nil, err
	}

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
			&r.ExternalAddress,
			&r.SlaveAddress,
			&r.IsRunning,
			&r.EstateName,
		)
		if err != nil {
			rows.Close()
			db.log.Error("Error in database query: ", err.Error())
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}
