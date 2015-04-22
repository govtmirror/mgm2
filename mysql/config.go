package mysql

import (
  "github.com/M-O-S-E-S/mgm2/core"
  "fmt"
  "database/sql"
  "github.com/satori/go.uuid"
)

func (db Database) GetDefaultConfigs() ([]core.ConfigOption, error) {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return nil, err}
  defer con.Close()

  cfgs := make([]core.ConfigOption, 0)

  rows, err := con.Query("SELECT section, item, content FROM iniConfig WHERE region IS NULL")
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  for rows.Next() {
    c := core.ConfigOption{}
    err = rows.Scan(
      &c.Section,
      &c.Item,
      &c.Content,
    )
    if err != nil {
      fmt.Println(err)
      return nil, err
    }
    cfgs = append(cfgs, c)
  }
  return cfgs, nil
}

func (db Database) GetConfigs(regionID uuid.UUID) ([]core.ConfigOption, error) {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return nil, err}
  defer con.Close()

  cfgs := make([]core.ConfigOption, 0)

  rows, err := con.Query("SELECT section, item, content FROM iniConfig WHERE region=?", regionID.String())
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  for rows.Next() {
    c := core.ConfigOption{}
    err = rows.Scan(
      &c.Section,
      &c.Item,
      &c.Content,
    )
    c.Region = regionID
    if err != nil {
      fmt.Println(err)
      return nil, err
    }
    cfgs = append(cfgs, c)
  }
  return cfgs, nil
}
