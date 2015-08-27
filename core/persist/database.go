package persist

import (
	"database/sql"
	"fmt"

	// load mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// Database is a mysql database interface for MGM
type Database struct {
	user     string
	password string
	database string
	host     string
}

// NewDatabase is a Database constructor
func NewDatabase(username string, password string, database string, host string) Database {
	return Database{username, password, database, host}
}

// TestConnection pings the database to confirm a valid connection
func (db Database) TestConnection() error {
	con, err := db.getConnection()
	if err != nil {
		return err
	}
	defer con.Close()

	err = con.Ping()
	return err
}

func (db Database) getConnection() (*sql.DB, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", db.user, db.password, db.host, db.database))
	if err != nil {
		return nil, err
	}
	return con, nil
}

// GetConnectionString retrieves an opensim style database connection string
func (db Database) GetConnectionString() string {
	return fmt.Sprintf("Data Source=%s;Database=%s;User ID=%s;Password=%s;Old Guids=true;", db.host, db.database, db.user, db.password)
}
