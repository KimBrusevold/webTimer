package database

type UserState int

const (
	Created           = 0
	Confirmed         = 1
	ResettingPasswrod = 2
)
