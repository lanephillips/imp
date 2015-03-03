package main

import (
	// "fmt"
	"github.com/go-sql-driver/mysql"
	"net/http"
	// "time"
)

type Host struct {
	HostId int64
	Name string
}

type Guest struct {
	GuestId int64
	Handle string
	HostId int64
	Token string
	CreatedDate mysql.NullTime
}

type UserHost struct {
	UserId int64
	HostId int64
	Token string
	CreatedDate mysql.NullTime
}

func GetUserGuestHandler(rw http.ResponseWriter, r *http.Request) {
	
}

func PutUserGuestHandler(rw http.ResponseWriter, r *http.Request) {
	
}

func PostGuestHandler(rw http.ResponseWriter, r *http.Request) {
	
}
