package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

func (r *TimerDB) ConfirmOneTimeCode(user User) error {
	if !user.OneTimeCode.Valid {
		log.Panic("User has no OneTimeCode set")
	}

	command := `SELECT id, password FROM users WHERE username = ? AND email = ? AND onetimecode= ?`
	row := r.db.QueryRow(command, user.Username, user.Email, user.OneTimeCode.String)

	var hashedPassword string
	err := row.Scan(&user.ID, &hashedPassword)
	if err != nil {
		log.Printf("Error when getting row values. %s", err)
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password)); err != nil {
		log.Printf("Passwords are not the same. %s", err)
		return err
	}

	command = "UPDATE users SET onetimecode = null, state = 1 WHERE id = ?;"
	_, err = r.db.Exec(command, user.ID)
	if err != nil {
		log.Printf("Could not update userstate. %s", err)
		return err
	}
	return nil
}

func (r *TimerDB) CreateUser(user User) (User, error) {
	uid := uuid.New()

	command := `SELECT id FROM users WHERE username = ? AND email = ?`

	password, error := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if error != nil {
		return User{}, error
	}
	row := r.db.QueryRow(command, user.Username, user.Email)

	err := row.Scan(&user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			command = `INSERT INTO users(username, email, password ,onetimecode, state)
			values(?, ?, ?, ?, ?)
			RETURNING id;`
			user.OneTimeCode = sql.NullString{
				String: uid.String(),
			}
			user.Password = string(password[:])
			row = r.db.QueryRow(command, user.Username, user.Email, user.Password, user.OneTimeCode.String, Created)

			err := row.Scan(&user.ID)
			if err != nil {
				return User{}, err
			}

			return user, nil
		}
		return User{}, err
	}
	return User{}, errors.New("user already exists")
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

	row := r.db.QueryRow(command, email)

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

func (r *TimerDB) SetNewOnetimeCode(username string, email string) (string, error) {
	command := `SELECT id FROM users WHERE email = ? and username = ?;`
	row := r.db.QueryRow(command, email, username)

	var id int64
	if err := row.Scan(&id); err != nil {
		return "", err
	}

	command = `UPDATE users SET 
		onetimecode = ?,
		authcode = NULL,
		state = 2
		WHERE id = ?;`

	uid := uuid.New().String()
	_, err := r.db.Exec(command, uid, id)
	if err != nil {
		return "", err
	}

	return uid, nil
}

func (r *TimerDB) UserAuthProcess(email string, password string) (*User, error) {

	command := `SELECT id, username, password FROM users WHERE email = ? AND status = 1;`

	row := r.db.QueryRow(command, email)

	user := User{}

	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
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

	res := r.db.QueryRow(`SELECT count(id) FROM times WHERE userid = ? AND endtime IS NULL`, userId)
	// if err != nil {
	// 	log.Printf("Noe gikk galt under spørring på tider")
	// 	log.Print(err.Error())
	// 	return err
	// }
	var n int
	err := res.Scan(&n)
	if err != nil {
		log.Printf("kunne ikke lese antall rader påvirket")
		return err
	}
	if n > 0 {
		log.Printf("Antall tider startet: %d", n)
		log.Print("Tid er allerede påbegynt")
		return nil
	}

	command := `INSERT INTO times(starttime, userid) values(?,?)`
	_, err = r.db.Exec(command, startTime, userId)
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

type RetrieveTimesResponse struct {
	Place        int
	Username     string
	ComputedTime int64
}

func (r *TimerDB) RetrieveTimes() ([]RetrieveTimesResponse, error) {
	query := `SELECT ROW_NUMBER () OVER (ORDER BY times.computedtime ASC) rownum, min(times.computedtime), username FROM times 
		INNER JOIN users on users.id = userid
		WHERE times.computedtime IS NOT NULL
		GROUP BY userid;`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var times []RetrieveTimesResponse

	for rows.Next() {
		var tim RetrieveTimesResponse
		if err := rows.Scan(&tim.Place, &tim.ComputedTime, &tim.Username); err != nil {
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
