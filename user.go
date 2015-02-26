package main

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"net/http"
	"time"
)

type User struct {
	UserId int64
	Handle string
	Status string
	Biography string
	Email string
	IsValidEmail bool
	EmailValidationToken string
	EmailValidationDate mysql.NullTime
	PasswordHash string
	JoinedDate mysql.NullTime
	IsDisabled bool
}

func PostUserHandler(rw http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	// fmt.Println("client ip is", ip)

	// rate limit new user creation by ip
	ipLimit, err := FetchIPLimit(db, ip)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	if ipLimit.UsersAllowedCount <= 0 && ipLimit.CountResetDate.Valid && time.Now().Before(ipLimit.CountResetDate.Time) {
		fmt.Println("Too many new user accounts from this address.", ipLimit)
		sendError(rw, 429, "Too many new user accounts from this address.")
		return
	}

	r.ParseForm()

	handle := r.PostFormValue("handle")
	if len(handle) == 0 {
		fmt.Println("Missing handle.")
    	sendError(rw, http.StatusBadRequest, "Missing handle.")
		return
	}

	email, err := mail.ParseAddress(r.PostFormValue("email"))
	if err != nil {
		fmt.Println(err)
    	sendError(rw, http.StatusBadRequest, "Invalid email address.")
		return
	}

    // TODO: check password strength
	password := r.PostFormValue("password")
	if len(password) == 0 {
		fmt.Println("Missing password.")
    	sendError(rw, http.StatusBadRequest, "Missing password.")
		return
	}

	// look up handle to see if this user already exists
    var count int64
    err = db.Get(&count, "SELECT COUNT(*) FROM User WHERE Handle LIKE ?", handle)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
    if count > 0 {
		fmt.Println("Handle already in use.")
    	sendError(rw, http.StatusConflict, "That handle is already in use.")
		return
    }

    err = db.Get(&count, "SELECT COUNT(*) FROM User WHERE Email LIKE ?", email.Address)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
    if count > 0 {
		fmt.Println("Email already in use.")
    	sendError(rw, http.StatusConflict, "That email address is already in use.")
		return
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	var u User
    u.Handle = handle
    u.Email = email.Address
    u.Status = ""
    u.Biography = ""
    u.PasswordHash = string(hash)
	fmt.Println(u)

	result, err := db.NamedExec("INSERT INTO `User` (`Handle`, `Status`, `Biography`, `Email`, `PasswordHash`) " +
			"VALUES (:Handle, :Status, :Biography, :Email, :PasswordHash)", &u)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	u.UserId, err = result.LastInsertId()
	if err != nil {
		fmt.Println(err)
	}

	// go ahead and log user in
	t, err := MakeToken(db, &u)
	if err != nil {
		fmt.Println(err)
		// something went wrong, but at least we created the user, so don't die here
	}
	ipLimit.LogNewUser(db)

	resp := map[string]interface{}{
		"user": &u,
		"token": t.Token,
	}

	sendData(rw, http.StatusCreated, resp)
}
