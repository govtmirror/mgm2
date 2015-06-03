package database

import (
	"database/sql"
	"fmt"

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
}

// Database is the database interface for persisting data
type Database interface {
	TestConnection() error

	GetConnection() (*sql.DB, error)

	GetConnectionString() string
}

// NewDatabase is a Database constructor
func NewDatabase(username string, password string, database string, host string) Database {
	return db{username, password, database, host}
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

func (db db) GetConnectionString() string {
	return fmt.Sprintf("Data Source=%s;Database=%s;User ID=%s;Password=%s;Old Guids=true;", db.host, db.database, db.user, db.password)
}
