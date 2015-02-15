package main

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"log"
)

type User struct {
	UserId int64
	Handle string
	Status string
	Biography string
	JoinedDate mysql.NullTime
}

func (u *User) Fetch(db *sql.DB, userId int64) {
	// TODO: 
}

// TODO: filters
func QueryUsers(db *sql.DB) (users []*User) {
	rows, err := db.Query("SELECT `UserId`, `Handle`, `Status`, `Biography`, `JoinedDate` FROM `User` WHERE 1")	// TODO: user id
	if err != nil {
	    log.Fatal(err)
	    // TODO: 500
	}

	users = make([]*User, 0)
	for rows.Next() {
	    var user User
	    if err := rows.Scan(&user.UserId, &user.Handle, &user.Status, &user.Biography, &user.JoinedDate); err != nil {
	        log.Fatal(err)
	    }
	    users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
	    log.Fatal(err)
	    // TODO: 500
	}
	return
}