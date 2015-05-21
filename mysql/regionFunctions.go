package mysql

import (
	"database/sql"
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
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning, EstateName, status from regions, estate_map, estate_settings " +
			"where estate_map.RegionID = regions.uuid AND estate_map.EstateID = estate_settings.EstateID")
	defer rows.Close()
	if err != nil {
		fmt.Println(err)
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
			&r.Status,
		)
		if err != nil {
			rows.Close()
			fmt.Println(err)
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}

// GetRegionsOnHost retrieves all region records for a specified host
func (db db) GetRegionsOnHost(host mgm.Host) ([]mgm.Region, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning, status from regions "+
			"where slaveAddress=?", host.Address)
	defer rows.Close()
	if err != nil {
		fmt.Println(err)
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
			&r.Status,
		)
		if err != nil {
			rows.Close()
			fmt.Println(err)
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
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning, EstateName, status from regions, estate_map, estate_settings " +
			"where estate_map.RegionID = regions.uuid AND estate_map.EstateID = estate_settings.EstateID AND uuid in " +
			"(SELECT RegionID FROM estate_map WHERE " +
			"EstateID in (SELECT EstateID FROM estate_settings WHERE EstateOwner=\"" + guid.String() + "\") OR " +
			"EstateID in (SELECT EstateID from estate_managers WHERE uuid=\"" + guid.String() + "\"))")
	defer rows.Close()
	if err != nil {
		fmt.Println(err)
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
			&r.Status,
		)
		if err != nil {
			rows.Close()
			fmt.Println(err)
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}
