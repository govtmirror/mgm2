package mysql

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/M-O-S-E-S/mgm/mgm"
	//mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// GetPendingUsers retrieves all pending users in mgm
func (db db) GetPendingUsers() ([]mgm.PendingUser, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", db.user, db.password, db.host, db.database))
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query("Select * from users")
	defer rows.Close()
	if err != nil {
		db.log.Error("Error in database query: ", err.Error())
		return nil, err
	}

	var users []mgm.PendingUser
	for rows.Next() {
		u := mgm.PendingUser{}
		err = rows.Scan(
			&u.Name,
			&u.Email,
			&u.Gender,
			&u.PasswordHash,
			&u.Registered,
			&u.Summary,
		)
		if err != nil {
			rows.Close()
			db.log.Error("Error in database query: ", err.Error())
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// AddPendingUser records a user registration for later approval
func (db db) AddPendingUser(name string, email string, template string, password string, summary string) error {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return err
	}
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

func (db db) IsEmailUnique(email string) (bool, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return false, err
	}
	defer con.Close()

	row := con.QueryRow("SELECT email FROM users WHERE email=?", email)
	var test string
	err = row.Scan(&test)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (db db) IsNameUnique(name string) (bool, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return false, err
	}
	defer con.Close()

	row := con.QueryRow("SELECT name FROM users WHERE name=?", name)
	var test string
	err = row.Scan(&test)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
