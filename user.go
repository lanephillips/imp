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
	Email string
	IsValidEmail bool
	EmailValidationToken string
	EmailValidationDate mysql.NullTime
	PasswordHash string
	JoinedDate mysql.NullTime
	IsDisabled bool
}

// TODO: not sure if this is the right way to search both handle and email
// TODO: we really need to use sqlx instead of this ORM style
func (u *User) Fetch(db *sql.DB, handleOrEmail string) (err error) {
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
