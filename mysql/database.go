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
  //regionManager RegionManager
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
func (db Database) GetAllRegions() (error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return err}
  defer con.Close()
  
  rows, err := con.Query("SELECT * FROM regions")
  for rows.Next() {
    r := core.Region{}
    err = rows.Scan(
      &r.UUID,
      &r.Name,
      &r.Size,
      &r.HttpPort,
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
      return err
    }
    //db.regionManager.LoadedRegion(r)
  }
  return nil
}