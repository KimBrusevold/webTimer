package timer

import "database/sql"

type Timer struct {
	ID        int64
	UserID    int64
	StartTime int64
	EndTime   int64
}

type User struct {
	ID          int64
	Username    string
	Email       string
	OneTimeCode sql.NullString
	authcode    sql.NullString
}
