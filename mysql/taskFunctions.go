package mysql

import (
  "fmt"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "github.com/M-O-S-E-S/mgm/core"
  "github.com/satori/go.uuid"
)

func (db Database) CreateTask(string, uuid.UUID, string) (core.Job, error) {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return core.Job{}, err}
  defer con.Close()

  return core.Job{}, nil
}
