package database

import (
	"database/sql"
	"fmt"

	"github.com/m-o-s-e-s/mgm/core"

	// load mysql driver
	_ "github.com/go-sql-driver/mysql"
)

//type RegionManager interface {
//	LoadedRegion(mgm.Region)
//}

type db struct {
	user     string
	password string
	database string
	host     string
	log      core.Logger
}

// Database is the database interface for persisting data
type Database interface {
	TestConnection() error

	GetConnection() (*sql.DB, error)
}

// NewDatabase is a Database constructor
func NewDatabase(username string, password string, database string, host string, log core.Logger) Database {
	return db{username, password, database, host, log}
}

func (db db) TestConnection() error {
	con, err := db.GetConnection()
	if err != nil {
		return err
	}
	defer con.Close()

	err = con.Ping()
	return err
}

func (db db) GetConnection() (*sql.DB, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", db.user, db.password, db.host, db.database))
	if err != nil {
		return nil, err
	}
	return con, nil
}
