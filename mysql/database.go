package mysql

import (
  "fmt"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "github.com/M-O-S-E-S/mgm2/core"
  "github.com/satori/go.uuid"
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

func (db Database) GetAllRegions() ([]core.Region, error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return nil, err}
  defer con.Close()

  rows, err := con.Query(
    "Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning, EstateName, status from regions, estate_map, estate_settings " +
    "where estate_map.RegionID = regions.uuid AND estate_map.EstateID = estate_settings.EstateID")

  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  regions := make([]core.Region, 0)
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

func (db Database) GetRegionsFor(guid uuid.UUID) ([]core.Region, error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return nil, err}
  defer con.Close()
  
  rows, err := con.Query(
    "Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning, EstateName, status from regions, estate_map, estate_settings " +
    "where estate_map.RegionID = regions.uuid AND estate_map.EstateID = estate_settings.EstateID AND uuid in " +
    "(SELECT RegionID FROM estate_map WHERE " +
    "EstateID in (SELECT EstateID FROM estate_settings WHERE EstateOwner=\"" + guid.String() + "\") OR " +
    "EstateID in (SELECT EstateID from estate_managers WHERE uuid=\"" + guid.String() + "\"))")

  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  regions := make([]core.Region, 0)
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

