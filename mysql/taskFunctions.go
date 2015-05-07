package mysql

import (
  "fmt"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "github.com/M-O-S-E-S/mgm/core"
  "github.com/satori/go.uuid"
)

func (db Database) GetJobsForUser(userID uuid.UUID) ([]core.Job, error) {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", db.user, db.password, db.host, db.database))
  if err != nil {return core.Job{}, err}
  defer con.Close()

}

func (db Database) CreateJob(taskType string, userID uuid.UUID, data string) (core.Job, error) {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", db.user, db.password, db.host, db.database))
  if err != nil {return core.Job{}, err}
  defer con.Close()

  res, err := con.Exec("INSERT INTO jobs (type, user, data) VALUES (?,?,?)", taskType, userID.String(), data)
  if err != nil {
    return core.Job{}, err
  }
  id, _ := res.LastInsertId()
  job := core.Job{}
  err = con.QueryRow("SELECT * FROM jobs WHERE id=?",id).Scan(&job.ID, &job.Timestamp, &job.Type, &job.User, &job.Data)
  if err != nil {
    return core.Job{}, err
  }

  return job, nil
}
