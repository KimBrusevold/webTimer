package timer

import (
	"database/sql"
	"errors"
)

var (
	ErrUpdateFailed = errors.New("Update failed")
)

type TimerDB struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *TimerDB {
	return &TimerDB{
		db: db,
	}
}

func (r *TimerDB) Migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS users(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        email TEXT NOT NULL
    );

	CREATE TABLE IF NOT EXISTS times(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		userid INTEGER,
        starttime INTEGER NOT NULL,
        endtime INTEGER NOT NULL,
		FOREIGN KEY(userid) REFERENCES users(id)
    );
    `

	_, err := r.db.Exec(query)
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
