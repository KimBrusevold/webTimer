package timer

import (
	"database/sql"
	"errors"
	"log"

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
		id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
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
        id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		userid bigint REFERENCES users (id),
        starttime bigint NOT NULL,
        endtime bigint NOT NULL,
		computedtime bigint
    );`

	_, err = r.db.Exec(query)

	if err != nil {
		return err
	}

	log.Print("Adding pgcrypto extention")
	query = `CREATE EXTENSION IF NOT EXISTS pgcrypto;`
	_, err = r.db.Exec(query)

	if err != nil {
		return err
	}

	log.Print("Add md5 function")
	query = `
	CREATE OR REPLACE FUNCTION md5(bytea) 
	RETURNS text AS $$ 
	SELECT encode(digest($1, 'md5'), 'hex')
	$$ LANGUAGE SQL STRICT IMMUTABLE;`
	_, err = r.db.Exec(query)

	if err != nil {
		return err
	}

	return err
}

func (r *TimerDB) CreateUser(user User) (int64, error) {
	uid := uuid.New()

	command := `SELECT id FROM users WHERE username = $1 AND email = md5($2)`
	row := r.db.QueryRow(command, user.Username, user.Email)
	var id int64

	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			command = `INSERT INTO users(username, email, onetimecode)
			values($1,md5($2), $3)
			RETURNING id;` //#d10ca8d11301c2f4993ac2279ce4b930

			row = r.db.QueryRow(command, user.Username, user.Email, uid.String())

			err := row.Scan(&id)
			if err != nil {
				return -1, err
			}

			return id, nil
		}
		return -1, err
	}

	command = `UPDATE users SET 
		onetimecode = $1,
		authcode = NULL
		WHERE id = $2;`

	_, err = r.db.Exec(command, uid.String(), id)

	if err != nil {
		return -1, err
	}

	return id, nil
}

func (r *TimerDB) GetUser(userid int64) (*User, error) {
	command := `SELECT id, username, email, onetimecode FROM users WHERE id = $1;`

	row := r.db.QueryRow(command, userid)

	user := User{}

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.OneTimeCode)
	if err != nil {
		return nil, err
	}

	return &user, err
}

func (r *TimerDB) UserAuthProcees(onetimeCode string) (*User, error) {
	command := `SELECT id, username, email, onetimecode FROM users WHERE onetimecode = $1;`

	row := r.db.QueryRow(command, onetimeCode)

	user := User{}

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.OneTimeCode)
	if err != nil {
		return nil, err
	}
	command = `UPDATE users SET 
		onetimecode = NULL,
		authcode = $1
		WHERE id = $2;`

	uid := uuid.New().String()
	_, err = r.db.Exec(command, uid, user.ID)
	if err != nil {
		return nil, err
	}
	user.Authcode.String = uid

	return &user, err
}

func (r *TimerDB) IsAuthorizedUser(authcode string, id int) bool {
	command := `SELECT id WHERE id = $1 AND authcode = $2`

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
