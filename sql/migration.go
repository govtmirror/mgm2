package sql

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/m-o-s-e-s/mgm/core/logger"
)

// Migrate tests the database version and performs any neccesary migrations
func (m MGMDB) Migrate(resourceFolder string) error {

	// list migration files we should have access to
	mgmFiles := map[int]string{
		0: "000-mgm.sql",
		1: "001-mgm.sql",
		2: "002-mgm.sql",
		3: "003-mgm.sql",
	}

	maxVersion := 3
	currentVersion := -1

	// read the current version in the database
	conn, err := m.db.getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	r := conn.QueryRow("SELECT MAX(version) AS current FROM mgmdb")
	_ = r.Scan(&currentVersion)
	m.log.Info("MGM is currently at version %v", currentVersion)
	if currentVersion == maxVersion {
		m.log.Info("MGM database is up to date")
		return nil
	}

	for i := currentVersion + 1; i <= maxVersion; i++ {
		m.log.Info("Migrating database to version %v", i)

		if i == 3 {
			err = migrate2to3(resourceFolder, conn, m.osdb, m.log)
			if err != nil {
				return err
			}
		}
		//read in sql file for next migration
		buf, err := ioutil.ReadFile(path.Join(resourceFolder, mgmFiles[i]))
		if err != nil {
			return err
		}

		// Golang can only process a single command at a time
		// here we assume that our input commands are terminated
		parts := strings.Split(string(buf), ";")
		for _, part := range parts {
			p := strings.TrimSpace(part)
			if p == "" {
				continue
			}
			_, err = conn.Exec(part)
			if err != nil {
				return err
			}
		}
	}

	m.log.Info("Database migration complete")

	return nil
}

func migrate2to3(resourceFolder string, conn *sql.DB, osdb Database, log logger.Log) error {
	log.Info("Migrating from before version 2, copying estate data to opensim database")
	//special case operation, we have estate tables that opensim uses in mgmdb
	//we need to move them into the opensim database for versions 3+

	buf, err := ioutil.ReadFile(path.Join(resourceFolder, "000-opensim.sql"))
	if err != nil {
		return err
	}

	osConn, err := osdb.getConnection()
	if err != nil {
		return err
	}
	defer osConn.Close()

	// Golang can only process a single command at a time
	// here we assume that our input commands are terminated
	parts := strings.Split(string(buf), ";")
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		_, err = osConn.Exec(part)
		if err != nil {
			return err
		}
	}

	// the tables now exist in osdb, copy the data
	tables := []string{
		"estateban",
		"estate_groups",
		"estate_managers",
		"estate_map",
		"estate_settings",
		"estate_users",
	}

	for _, table := range tables {
		rows, err := conn.Query(fmt.Sprintf("Select * from %v", table))
		if err != nil {
			return err
		}
		var fields []interface{}
		cols, err := rows.Columns()
		if err != nil {
			return err
		}
		colValues := []string{}
		for _ = range cols {
			fields = append(fields, new(string))
			colValues = append(colValues, "?")
		}

		for rows.Next() {
			//scan the data from this row
			err = rows.Scan(fields...)
			if err != nil {
				return err
			}

			//insert it into this table here
			_, err = osConn.Exec(fmt.Sprintf(
				"INSERT IGNORE INTO %v (%v) VALUES (%v)",
				table,
				strings.Join(cols, ","),
				strings.Join(colValues, ","),
			), fields...)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
