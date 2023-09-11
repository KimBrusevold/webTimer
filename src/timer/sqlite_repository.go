package timer

import (
	"database/sql"
	"errors"
	"log"
	
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
        username TEXT NOT NULL,
        email TEXT NOT NULL
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
        endtime bigint NOT NULL
    );`

	_, err = r.db.Exec(query)

	return err
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
