package timer

import "database/sql"

type Timer struct {
	ID           int64
	UserID       int64
	StartTime    int64
	EndTime      int64
	ComputedTime sql.NullInt64
}

type User struct {
	ID          int64
	Username    string
	Email       string
	Password    string
	OneTimeCode sql.NullString
	Authcode    sql.NullString
}
