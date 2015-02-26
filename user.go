package main

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/mail"
	"net/http"
	"time"
	"github.com/jmoiron/sqlx"
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

func PostUser(rw http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	// fmt.Println("client ip is", ip)

	// rate limit new user creation by ip
	var ipLimit IPLimit
	err := ipLimit.Fetch(db, ip)
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
    u.UserId = -1
    u.Handle = handle
    u.Email = email.Address
    u.Status = ""
    u.Biography = ""
    u.PasswordHash = string(hash)
	fmt.Println(u)

    err = u.Save(db)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	// go ahead and log user in
	var t UserToken
	t.UserId = u.UserId

	err = t.Save(db)
	if err != nil {
		fmt.Println(err)
		// something went wrong, but at least we created the user, so don't die here
	}
	ipLimit.LogNewUser(db)

	resp := map[string]interface{}{
		"user": u,
		"token": t.Token,
	}

	sendData(rw, http.StatusCreated, resp)
}

// TODO: not sure if this is the right way to search both handle and email
// TODO: we really need to use sqlx instead of this ORM style
func (u *User) Fetch(db *sqlx.DB, handleOrEmail string) (err error) {
    u.UserId = -1

	stmt, err := db.Prepare("SELECT `UserId`, `Handle`, `Status`, `Biography`, `PasswordHash`, `JoinedDate` FROM `User` " +
		"WHERE Handle LIKE ? OR Email LIKE ? LIMIT 1")
	if err != nil {
	    log.Println(err)
	    return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(handleOrEmail, handleOrEmail)
	if err != nil {
	    log.Println(err)
	    return err
	}

	if rows.Next() {
	    if err := rows.Scan(&u.UserId, &u.Handle, &u.Status, &u.Biography, &u.PasswordHash, &u.JoinedDate); err != nil {
	        log.Println(err)
	    }
	}
	if err := rows.Err(); err != nil {
	    log.Println(err)
	}
	return err
}

func (u *User) Save(db *sqlx.DB) (err error) {
	if u.UserId > 0 {
		stmt, err := db.Prepare("UPDATE `User` SET `Handle` = ?, `Status` = ?, `Biography` = ?, `Email` = ?, `PasswordHash` = ? " +
			"WHERE UserId = ?")
		if err != nil {
		    log.Println(err)
		    return err
		}
		defer stmt.Close()

		result, err := stmt.Exec(u.Handle, u.Status, u.Biography, u.Email, u.PasswordHash, u.UserId)
        if err != nil {
		    log.Println(err)
		    return err
        }

        count, err := result.RowsAffected()
        if err != nil {
		    log.Println(err)
		    return err
        }
        if count != 1 {
        	log.Println("Expected to update 1 row, not %d", count)
        }
	} else {
		stmt, err := db.Prepare("INSERT INTO `User` (`Handle`, `Status`, `Biography`, `Email`, `PasswordHash`) " +
			"VALUES (?, ?, ?, ?, ?)")
		if err != nil {
		    log.Println(err)
		    return err
		}
		defer stmt.Close()

		result, err := stmt.Exec(u.Handle, u.Status, u.Biography, u.Email, u.PasswordHash)
        if err != nil {
		    log.Println(err)
		    return err
        }

        id, err := result.LastInsertId()
        if err != nil {
		    log.Println(err)
		    return err
        }
        u.UserId = id
	}
	return nil
}
