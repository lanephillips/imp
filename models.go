package main

import "github.com/go-sql-driver/mysql"

type User struct {
	UserId int64
	Handle string
	Status string
	Biography string
	JoinedDate mysql.NullTime
}
