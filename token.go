package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"time"
	"crypto/rand"
	"github.com/jmoiron/sqlx"
)

type UserToken struct {
	Token string
	UserId int64
	LoginTime mysql.NullTime
	LastSeenTime mysql.NullTime
}

func PostTokenHandler(rw http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	handleOrEmail := r.PostFormValue("handleOrEmail")
	if len(handleOrEmail) == 0 {
		fmt.Println("Missing handle or email.")
    	sendError(rw, http.StatusBadRequest, "Missing handle or email.")
		return
	}

	password := r.PostFormValue("password")
	if len(password) == 0 {
		fmt.Println("Missing password.")
    	sendError(rw, http.StatusBadRequest, "Missing password.")
		return
	}

	ip := getIP(r)
	// fmt.Println("client ip is", ip)

	// rate limit by ip
	ipLimit, err := FetchIPLimit(db, ip)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	if ipLimit.LastLoginAttemptDate.Valid && time.Now().Before(ipLimit.LastLoginAttemptDate.Time.Add(time.Duration(SecondsBetweenLoginAttemptsPerIP) * time.Second)) {
		fmt.Println("Too many login attempts from this address.", ipLimit)
		sendError(rw, 429, "Too many login attempts from this address.")
		return
	}

	// rate limit by handle even if it's not a real handle, because otherwise we would reveal its existence
	limit, err := FetchHandleLimit(db, handleOrEmail)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	if limit.LoginAttemptCount > 0 && time.Now().Before(limit.LastAttemptDate.Time.Add(time.Duration(limit.NextLoginDelay) * time.Second)) {
		fmt.Println("Too many login attempts.", limit)
		sendError(rw, 429, "Too many login attempts.")
		return
	}

    var u User
    err = db.Get(&u, "SELECT `UserId`, `Handle`, `Status`, `Biography`, `PasswordHash`, `JoinedDate` FROM `User` " +
		"WHERE Handle LIKE ? OR Email LIKE ? LIMIT 1", handleOrEmail, handleOrEmail)
    if err == sql.ErrNoRows {
    	// in order to prevent not found user failing more quickly than bad password
    	// proceed with checking password against dummy hash

	    // dummy, _ := bcrypt.GenerateFromPassword([]byte(RandomString(50)), bcrypt.DefaultCost)
	    // fmt.Println("dummy hash: ", string(dummy))
    	u.PasswordHash = "$2a$10$tg.SM/VMqShumLh/uhB1BOCFcQyCIBu4XvBf7lszBw2lMew1ubNWq"
    	u.UserId = -1
    } else if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}


    err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil || u.UserId <= 0 {
    	err = limit.Bump(db)
		if err != nil {
			fmt.Println(err)
		}
    	err = ipLimit.LogAttempt(db)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("Bad credentials.")
    	sendError(rw, http.StatusUnauthorized, "No user was found that matched the handle or email and password given.")
		return
	}
	limit.Clear(db)

	t, err := MakeToken(db, &u)
	if err != nil {
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	resp := map[string]interface{}{
		"user": &u,
		"token": t.Token,
	}

	sendData(rw, http.StatusCreated, resp)
}

func DeleteTokenHandler(rw http.ResponseWriter, r *http.Request) {
	err := DeleteToken(db, mux.Vars(r)["token"])
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	// TODO: is it silly to send "No Content" along with an evelope?
	sendData(rw, http.StatusNoContent, "")
}

func FetchToken(db *sqlx.DB, r *http.Request)  (*UserToken, error) {
	auth := r.Header.Get("Authorization")
	authPrefix := "IMP auth="
	if !strings.HasPrefix(auth, authPrefix) {
		return nil, nil
	}
	token := auth[len(authPrefix):]

	t := new(UserToken)
	err := db.Get(t, "SELECT `Token`, `UserId`, `LoginTime`, `LastSeenTime` FROM `UserToken` WHERE Token LIKE ? LIMIT 1", token)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
	    log.Println(err)
	    return nil, err
	}
	return t, nil
}

func MakeToken(db *sqlx.DB, user *User) (*UserToken, error) {
	t := new(UserToken)
	t.Token = RandomString(50)
	t.UserId = user.UserId
	t.LoginTime.Time = time.Now()
	t.LoginTime.Valid = true
	t.LastSeenTime.Time = time.Now()
	t.LastSeenTime.Valid = true

	_, err := db.NamedExec("INSERT INTO `UserToken` (`Token`, `UserId`, `LoginTime`, `LastSeenTime`) " +
			"VALUES (:Token, :UserId, :LoginTime, :LastSeenTime)", t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func DeleteToken(db *sqlx.DB, token string) (err error) {
	_, err = db.Exec("DELETE FROM `UserToken` WHERE Token LIKE ?", token)
	return
}

func RandomString(strSize int) string {
	dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
	     bytes[k] = dictionary[v % byte(len(dictionary))]
	}
	return string(bytes)
}
