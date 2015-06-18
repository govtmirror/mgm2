package user

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"

	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type userDatabase struct {
	mysql persist.Database
}

// GetPendingUsers retrieves all pending users in mgm
func (db userDatabase) GetPendingUsers() ([]mgm.PendingUser, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query("Select * from users")
	defer rows.Close()
	if err != nil {
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
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// AddPendingUser records a user registration for later approval
func (db userDatabase) AddPendingUser(name string, email string, template string, password string, summary string) error {
	con, err := db.mysql.GetConnection()
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

func (db userDatabase) IsEmailUnique(email string) (bool, error) {
	con, err := db.mysql.GetConnection()
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

func (db userDatabase) IsNameUnique(name string) (bool, error) {
	con, err := db.mysql.GetConnection()
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

func (db userDatabase) ScrubPasswordToken(token uuid.UUID) error {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Exec("DELETE FROM jobs WHERE data=?", token.String())
	if err != nil {
		return err
	}
	return nil
}

func (db userDatabase) ExpirePasswordTokens() error {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Exec("DELETE FROM jobs WHERE type=\"password_reset\" AND timestamp >= DATE_SUB(NOW(), INTERVAL 1 DAY)")
	if err != nil {
		return err
	}
	return nil
}

func (db userDatabase) CreatePasswordResetToken(userID uuid.UUID) (uuid.UUID, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return uuid.UUID{}, err
	}
	defer con.Close()

	token := uuid.NewV4()
	_, err = con.Exec("INSERT INTO jobs (type, user, data) VALUES(\"password_reset\", ?, ?)", userID.String(), token.String())
	if err != nil {
		return uuid.UUID{}, err
	}
	return token, nil
}

func (db userDatabase) ValidatePasswordToken(userID uuid.UUID, token uuid.UUID) (bool, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return false, err
	}
	defer con.Close()

	rows, err := con.Query("SELECT data FROM jobs WHERE type=\"password_reset\" AND user=? AND timestamp >= DATE_SUB(NOW(), INTERVAL 1 DAY)", userID.String())
	defer rows.Close()
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
