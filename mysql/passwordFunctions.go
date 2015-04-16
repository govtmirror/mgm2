package mysql

import (
  "fmt"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "github.com/satori/go.uuid"
)


func (db Database) ScrubPasswordToken(token uuid.UUID) error {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return err}
  defer con.Close()

  _, err = con.Exec("DELETE FROM jobs WHERE data=?", token.String())
  if err != nil {
    return err
  }
  return nil
}

func (db Database) ExpirePasswordTokens() error {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return err}
  defer con.Close()

  _, err = con.Exec("DELETE FROM jobs WHERE type=\"password_reset\" AND timestamp >= DATE_SUB(NOW(), INTERVAL 1 DAY)")
  if err != nil {
    return err
  }
  return nil
}

func (db Database) CreatePasswordResetToken(userID uuid.UUID) (uuid.UUID, error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return uuid.UUID{}, err}
  defer con.Close()

  token := uuid.NewV4()
  _, err = con.Exec("INSERT INTO jobs (type, user, data) VALUES(\"password_reset\", ?, ?)", userID.String(), token.String())
  if err != nil {
    return uuid.UUID{}, err
  }
  return token, nil
}

func (db Database) ValidatePasswordToken(userID uuid.UUID, token uuid.UUID) (bool, error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return false, err}
  defer con.Close()

  rows, err := con.Query("SELECT data FROM jobs WHERE type=\"password_reset\" AND user=? AND timestamp >= DATE_SUB(NOW(), INTERVAL 1 DAY)", userID.String())
  if err != nil {
    return false, err
  }
  for rows.Next() {
    var scanToken uuid.UUID
    err = rows.Scan(&scanToken)
    if err != nil {
      return false, err
    }
    if scanToken == token {
      return true, nil
    }
  }
  return false, nil
}
