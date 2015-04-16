package mysql

import (
  "fmt"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "github.com/M-O-S-E-S/mgm2/core"
)

type RegionManager interface {
  LoadedRegion(core.Region)
}

type Database struct {
  user string
  password string
  database string
  host string
}

func NewDatabase(username string, password string, database string, host string) *Database{
  return &Database{username, password, database, host}
}

func (db Database) TestConnection() error {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return err}
  defer con.Close()
  
  err = con.Ping()
  return err
}
