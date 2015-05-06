package mysql

import (
  "github.com/M-O-S-E-S/mgm/core"
  "fmt"
  "database/sql"
  "github.com/satori/go.uuid"
)


func (db Database) GetEstates() ([]core.Estate, error) {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return nil, err}
  defer con.Close()

  estates := make([]core.Estate, 0)

  rows, err := con.Query("Select EstateID, EstateName, EstateOwner from estate_settings")
  defer rows.Close()
  for rows.Next() {
    e := core.Estate{Managers: make([]uuid.UUID,0), Regions: make([]uuid.UUID,0)}
    err = rows.Scan(
      &e.ID,
      &e.Name,
      &e.Owner,
    )
    if err != nil {
      fmt.Println(err)
      return nil, err
    }
    estates = append(estates, e)
  }

  for i, e := range estates {
    //lookup managers
    rows, err := con.Query("SELECT uuid FROM estate_managers WHERE EstateID=?", e.ID)
    defer rows.Close()
    if err != nil {
      fmt.Println(err)
      return nil, err
    }
    for rows.Next() {
      guid := uuid.UUID{}
      err = rows.Scan(&guid)
      if err != nil {
        fmt.Println(err)
        return nil, err
      }
      estates[i].Managers = append(estates[i].Managers, guid)
    }
    //lookup regions
    rows, err = con.Query("SELECT RegionID FROM estate_map WHERE EstateID=?", e.ID)
    defer rows.Close()
    if err != nil {
      fmt.Println(err)
      return nil, err
    }
    for rows.Next() {
      guid := uuid.UUID{}
      err = rows.Scan(&guid)
      if err != nil {
        fmt.Println(err)
        return nil, err
      }
      estates[i].Regions = append(estates[i].Regions, guid)
    }
  }

  return estates, nil
}
