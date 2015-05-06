package mysql

import (
  "fmt"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "github.com/M-O-S-E-S/mgm/core"
  "github.com/satori/go.uuid"
)

func (db Database) GetRegions() ([]core.Region, error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return nil, err}
  defer con.Close()

  rows, err := con.Query(
    "Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning, EstateName, status from regions, estate_map, estate_settings " +
    "where estate_map.RegionID = regions.uuid AND estate_map.EstateID = estate_settings.EstateID")
  defer rows.Close()
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

func (db Database) GetRegionsOnHost(address string) ([]core.Region, error) {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return nil, err}
  defer con.Close()

  rows, err := con.Query(
    "Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning, status from regions " +
    "where slaveAddress=?", address)
  defer rows.Close()
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

func (db Database) GetRegionsForUser(guid uuid.UUID) ([]core.Region, error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return nil, err}
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
