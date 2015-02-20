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
	PasswordHash string
	JoinedDate mysql.NullTime
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

// TODO: filters
func QueryUsers(db *sql.DB) (users []*User, err error) {
	rows, err := db.Query("SELECT `UserId`, `Handle`, `Status`, `Biography`, `JoinedDate` FROM `User` WHERE 1")	// TODO: user id
	if err != nil {
	    log.Println(err)
	    return nil, err
	}

	users = make([]*User, 0)
	for rows.Next() {
	    var user User
	    if err := rows.Scan(&user.UserId, &user.Handle, &user.Status, &user.Biography, &user.JoinedDate); err != nil {
	        log.Println(err)
	    }
	    users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
	    log.Println(err)
	}
	return users, err
}