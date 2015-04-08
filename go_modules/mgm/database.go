package mgm

import (
  "fmt"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

type database struct {
  user string
  password string
  database string
  host string
  rMgr regionManager
}

func (db database) testConnection() error {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return err}
  defer con.Close()
  
  err = con.Ping()
  return err
}
func (db database) loadRegions() (error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return err}
  defer con.Close()
  
  rows, err := con.Query("SELECT * FROM regions")
  for rows.Next() {
    r := region{}
    err = rows.Scan(
      &r.uuid,
      &r.name,
      &r.size,
      &r.httpPort,
      &r.consolePort,
      &r.consoleUname,
      &r.consolePass,
      &r.locX,
      &r.locY,
      &r.externalAddress,
      &r.slaveAddress,
      &r.isRunning,
      &r.status,
    )
    if err != nil {
      rows.Close()
      fmt.Println(err)
      return err
    }
    db.rMgr.newRegions<-r
  }
  return nil
}