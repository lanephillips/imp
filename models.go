package main

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"log"
	"time"
	"crypto/rand"
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

type UserToken struct {
	Token string
	UserId int64
	LoginTime mysql.NullTime
	LastSeenTime mysql.NullTime
}

func (u *User) Fetch(db *sql.DB, handle string) (err error) {
    u.UserId = -1

	stmt, err := db.Prepare("SELECT `UserId`, `Handle`, `Status`, `Biography`, `JoinedDate` FROM `User` WHERE Handle LIKE ? LIMIT 1")
	if err != nil {
	    log.Println(err)
	    return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(handle)
	if err != nil {
	    log.Println(err)
	    return err
	}

	if rows.Next() {
	    if err := rows.Scan(&u.UserId, &u.Handle, &u.Status, &u.Biography, &u.JoinedDate); err != nil {
	        log.Println(err)
	    }
	}
	if err := rows.Err(); err != nil {
	    log.Println(err)
	}
	return err
}

func (u *User) Save(db *sql.DB) (err error) {
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

// TODO: DRY out this repeated fetch and save code, sqlx might help
func (t *UserToken) Fetch(db *sql.DB, token string) (err error) {
    t.Token = ""

	stmt, err := db.Prepare("SELECT `Token`, `UserId`, `LoginTime`, `LastSeenTime` FROM `UserToken` WHERE Token LIKE ? LIMIT 1")
	if err != nil {
	    log.Println(err)
	    return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(token)
	if err != nil {
	    log.Println(err)
	    return err
	}

	if rows.Next() {
	    if err := rows.Scan(&t.Token, &t.UserId, &t.LoginTime, &t.LastSeenTime); err != nil {
	        log.Println(err)
	    }
	}
	if err := rows.Err(); err != nil {
	    log.Println(err)
	}
	return err
}

func (t *UserToken) Save(db *sql.DB) (err error) {
	if len(t.Token) > 0 {
		stmt, err := db.Prepare("UPDATE `UserToken` SET `LastSeenTime` = ? WHERE Token LIKE ?")
		if err != nil {
		    log.Println(err)
		    return err
		}
		defer stmt.Close()

		t.LastSeenTime.Time = time.Now()
		t.LastSeenTime.Valid = true

		result, err := stmt.Exec(t.LastSeenTime, t.Token)
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
		stmt, err := db.Prepare("INSERT INTO `UserToken` (`Token`, `UserId`, `LoginTime`, `LastSeenTime`) " +
			"VALUES (?, ?, ?, ?)")
		if err != nil {
		    log.Println(err)
		    return err
		}
		defer stmt.Close()

		t.Token = RandomString(50)
		t.LoginTime.Time = time.Now()
		t.LoginTime.Valid = true
		t.LastSeenTime.Time = time.Now()
		t.LastSeenTime.Valid = true

		_, err = stmt.Exec(t.Token, t.UserId, t.LoginTime, t.LastSeenTime)
        if err != nil {
		    log.Println(err)
		    return err
        }
	}
	return nil
}

func (t *UserToken) Delete(db *sql.DB) (err error) {
	if len(t.Token) > 0 {
		stmt, err := db.Prepare("DELETE FROM `UserToken` WHERE Token LIKE ?")
		if err != nil {
		    log.Println(err)
		    return err
		}
		defer stmt.Close()

		result, err := stmt.Exec(t.Token)
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
        t.Token = ""
	} else {
		// TODO: error
	}
	return nil
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
