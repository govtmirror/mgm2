package user

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"

	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type userDatabase struct {
	mysql database.Database
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

// GetEstates retrieves all estates from mgm
func (db userDatabase) GetEstates() ([]mgm.Estate, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	var estates []mgm.Estate

	rows, err := con.Query("Select EstateID, EstateName, EstateOwner from estate_settings")
	defer rows.Close()
	for rows.Next() {
		e := mgm.Estate{Managers: make([]uuid.UUID, 0), Regions: make([]uuid.UUID, 0)}
		err = rows.Scan(
			&e.ID,
			&e.Name,
			&e.Owner,
		)
		if err != nil {
			return nil, err
		}
		estates = append(estates, e)
	}

	for i, e := range estates {
		//lookup managers
		rows, err := con.Query("SELECT uuid FROM estate_managers WHERE EstateID=?", e.ID)
		defer rows.Close()
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			guid := uuid.UUID{}
			err = rows.Scan(&guid)
			if err != nil {
				return nil, err
			}
			estates[i].Managers = append(estates[i].Managers, guid)
		}
		//lookup regions
		rows, err = con.Query("SELECT RegionID FROM estate_map WHERE EstateID=?", e.ID)
		defer rows.Close()
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			guid := uuid.UUID{}
			err = rows.Scan(&guid)
			if err != nil {
				return nil, err
			}
			estates[i].Regions = append(estates[i].Regions, guid)
		}
	}

	return estates, nil
}
