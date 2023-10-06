package timer

import (
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUpdateFailed = errors.New("Update failed")
)

type TimerDB struct {
	db *sql.DB
}

func NewDbTimerRepository(db *sql.DB) *TimerDB {
	return &TimerDB{
		db: db,
	}
}

func (r *TimerDB) Migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS users(
		id INTEGER NOT NULL PRIMARY KEY,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
        onetimecode TEXT,
		authcode TEXT
		);
		`

	log.Print("Creating users table if not exists")
	_, err := r.db.Exec(query)

	if err != nil {
		return err
	}

	log.Print("Creating times table if not exists")
	query = `
	CREATE TABLE IF NOT EXISTS times(
        id INTEGER NOT NULL PRIMARY KEY,
		userid INTEGER REFERENCES users (id),
        starttime INTEGER NOT NULL,
        endtime INTEGER NULL,
		computedtime INTEGER
    );`

	_, err = r.db.Exec(query)

	if err != nil {
		return err
	}

	return err
}

func (r *TimerDB) CreateUser(user User) (int64, error) {
	uid := uuid.New()

	command := `SELECT id FROM users WHERE username = ? AND email = ?`
	md5Email := fmt.Sprintf("%x", md5.Sum([]byte(user.Email)))

	row := r.db.QueryRow(command, user.Username, md5Email)
	var id int64

	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			command = `INSERT INTO users(username, email, onetimecode)
			values(?, ?, ?)
			RETURNING id;`

			row = r.db.QueryRow(command, user.Username, md5Email, uid.String())

			err := row.Scan(&id)
			if err != nil {
				return -1, err
			}

			return id, nil
		}
		return -1, err
	}

	command = `UPDATE users SET 
		onetimecode = ?,
		authcode = NULL
		WHERE id = ?;`

	_, err = r.db.Exec(command, uid.String(), id)

	if err != nil {
		return -1, err
	}

	return id, nil
}

func (r *TimerDB) GetUser(userid int64) (*User, error) {
	command := `SELECT id, username, email, onetimecode FROM users WHERE id = ?;`

	row := r.db.QueryRow(command, userid)

	user := User{}

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.OneTimeCode)
	if err != nil {
		return nil, err
	}

	return &user, err
}

func (r *TimerDB) UserExistsWithUsername(username string) (bool, error) {
	command := `SELECT id FROM users WHERE username = ?`
	row := r.db.QueryRow(command, username)

	var id int64
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (r *TimerDB) UserExistsWithEmail(email string) (bool, int64, error) {
	command := `SELECT id FROM users WHERE email = ?`
	md5String := fmt.Sprintf("%x", md5.Sum([]byte(email)))

	row := r.db.QueryRow(command, md5String)

	var id int64
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, -1, nil
		} else {
			return false, -1, err
		}
	}
	return true, id, nil
}

func (r *TimerDB) SetNewOnetimeCode(email string) (string, error) {
	command := `SELECT id FROM users WHERE email = ?;`
	md5String := fmt.Sprintf("%x", md5.Sum([]byte(email)))
	row := r.db.QueryRow(command, md5String)

	var id int64
	if err := row.Scan(&id); err != nil {
		return "", err
	}

	command = `UPDATE users SET 
		onetimecode = ?,
		authcode = NULL
		WHERE id = ?;`

	uid := uuid.New().String()
	_, err := r.db.Exec(command, uid, id)
	if err != nil {
		return "", err
	}

	return uid, nil
}

func (r *TimerDB) UserAuthProcess(onetimeCode string) (*User, error) {
	command := `SELECT id, username, email, onetimecode FROM users WHERE onetimecode = ?;`

	row := r.db.QueryRow(command, onetimeCode)

	user := User{}

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.OneTimeCode)
	if err != nil {
		return nil, err
	}
	command = `UPDATE users SET 
		onetimecode = NULL,
		authcode = ?
		WHERE id = ?;`

	uid := uuid.New().String()
	_, err = r.db.Exec(command, uid, user.ID)
	if err != nil {
		return nil, err
	}
	user.Authcode.String = uid

	return &user, err
}

func (r *TimerDB) IsAuthorizedUser(authcode string, id int) bool {
	command := `SELECT id WHERE id = ? AND authcode = ?`

	row := r.db.QueryRow(command, id, authcode)
	var resid int64
	err := row.Scan(&resid)

	return err != nil
}

func (r *TimerDB) Create(timer Timer) (*Timer, error) {
	res, err := r.db.Exec("INSERT INTO times(starttime, endtime) values(?,?)", timer.StartTime, timer.EndTime)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	timer.ID = id

	return &timer, nil
}

func (r *TimerDB) StartTimer(userId int) error {
	startTime := time.Now().UnixMilli()
	command := `INSERT INTO times(starttime, userid) values(?,?)`
	_, err := r.db.Exec(command, startTime, userId)
	if err != nil {
		return err
	}
	return nil
}
func (r *TimerDB) EndTimeTimer(userId int) (int64, error) {
	query := `SELECT id, starttime FROM times WHERE userid = ? AND endtime IS NULL`
	row := r.db.QueryRow(query, userId)

	var id int64
	var startTime int64
	err := row.Scan(&id, &startTime)
	if err != nil {
		return -1, err
	}

	endtime := time.Now().UnixMilli()
	computed := endtime - startTime
	_, err = r.db.Exec("UPDATE times SET endtime = ?, computedtime = ? WHERE id = ?", endtime, computed, id)

	return computed, err
}

func (r *TimerDB) RetrieveTimes() ([]Timer, error) {
	query := `SELECT id, userid, computedtime FROM times 
		WHERE computedtime IS NOT NULL 
		ORDER BY computedtime ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var times []Timer

	for rows.Next() {
		var tim Timer
		if err := rows.Scan(&tim.ID, &tim.UserID, &tim.ComputedTime); err != nil {
			return times, err
		}

		times = append(times, tim)
	}

	if err = rows.Err(); err != nil {
		return times, err
	}

	return times, nil
}

func (r *TimerDB) Update(id int64, updated Timer) (*Timer, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec("UPDATE times SET endtime = ? WHERE id = ?", updated.EndTime, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUpdateFailed
	}

	return &updated, nil
}
