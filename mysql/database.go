package mysql

import (
  "fmt"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "github.com/M-O-S-E-S/mgm2/core"
  "github.com/satori/go.uuid"
  "crypto/md5"
  "encoding/hex"
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

func (db Database) AddPendingUser(name string, email string, template string, password string, summary string) error {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return err}
  defer con.Close()

  hasher := md5.New()
  hasher.Write([]byte(password))
  creds := hex.EncodeToString(hasher.Sum(nil))

  _, err = con.Exec("INSERT INTO users (name, email, gender, password, summary) VALUES(?, ?, ?, ?, ?)",
                   name, email, template, creds, summary)
  if err != nil {
    return err
  }
  return nil
}

func (db Database) IsEmailUnique(email string) (bool, error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return false, err}
  defer con.Close()

  row := con.QueryRow("SELECT email FROM users WHERE email=?", email)
  var test string
  err = row.Scan(&test)
  if err != nil{
    if err == sql.ErrNoRows {
      return true, nil
    }
    return false, err
  }
  return false, nil
}

func (db Database) IsNameUnique(name string) (bool, error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return false, err}
  defer con.Close()

  row := con.QueryRow("SELECT name FROM users WHERE name=?", name)
  var test string
  err = row.Scan(&test)
  if err != nil{
    if err == sql.ErrNoRows {
      return true, nil
    }
    return false, err
  }
  return false, nil
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

