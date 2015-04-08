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
}

func (st database) testConnection() error {
  db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@%v/%v", st.user, st.password, st.host, st.database))
  if err != nil {
    return err
  }
  err = db.Ping()
  return err
}
func (st database) getRegions() ([]region, error){
  //db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@%v/%v", st.user, st.password, st.host, st.database))
  //if err != nil {
  //  return nil, err
  //}
  return nil, nil
}