package main

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"log"
	"time"
	"github.com/jmoiron/sqlx"
)

type HandleLimit struct {
	Handle string
	LoginAttemptCount int64
	LastAttemptDate mysql.NullTime
	NextLoginDelay int64
}

type IPLimit struct {
	IP string
	LastLoginAttemptDate mysql.NullTime
	UsersAllowedCount int64
	CountResetDate mysql.NullTime
}

const (
	NewUsersPerIPPerDay = 24
	SecondsBetweenLoginAttemptsPerIP = 1
)

func FetchHandleLimit(db *sqlx.DB, handle string) (h *HandleLimit, err error) {
	h = new(HandleLimit)
    h.Handle = handle

	err = db.Get(h, "SELECT `LoginAttemptCount`, `LastAttemptDate`, `NextLoginDelay` FROM `HandleLimit` WHERE `Handle` LIKE ?", handle)
	if err == sql.ErrNoRows {
		h.NextLoginDelay = 1
		return h, nil
	} else if err != nil {
	    log.Println(err)
	    return nil, err
	}
	return h, nil
}

func (h *HandleLimit) Bump(db *sqlx.DB) (err error) {
	h.LoginAttemptCount = 1
	h.LastAttemptDate.Time = time.Now()
	h.LastAttemptDate.Valid = true
	h.NextLoginDelay = 1

	_, err = db.NamedExec("INSERT INTO `HandleLimit` (`Handle`, `LoginAttemptCount`, `LastAttemptDate`, `NextLoginDelay`) " +
		"VALUES (:Handle, :LoginAttemptCount, :LastAttemptDate, :NextLoginDelay) " +
		"ON DUPLICATE KEY UPDATE `LoginAttemptCount` = `LoginAttemptCount` + 1, `NextLoginDelay` = 2 * `NextLoginDelay`, " +
		"`LastAttemptDate` = VALUES(`LastAttemptDate`)", h)
	if err != nil {
	    log.Println(err)
	    return err
	}
	return nil
}

func (h *HandleLimit) Clear(db *sqlx.DB) (err error) {
	_, err = db.Exec("DELETE FROM `HandleLimit` WHERE Handle LIKE ?", h.Handle)
	if err != nil {
	    log.Println(err)
	    return err
	}
	return nil
}

func (h *IPLimit) Fetch(db *sqlx.DB, ip string) (err error) {
    h.IP = ip
    h.UsersAllowedCount = NewUsersPerIPPerDay

	stmt, err := db.Prepare("SELECT `LastLoginAttemptDate`, `UsersAllowedCount`, `CountResetDate` FROM `IPLimit` WHERE `IP` LIKE ?")
	if err != nil {
	    log.Println(err)
	    return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(ip)
	if err != nil {
	    log.Println(err)
	    return err
	}

	if rows.Next() {
	    if err := rows.Scan(&h.LastLoginAttemptDate, &h.UsersAllowedCount, &h.CountResetDate); err != nil {
	        log.Println(err)
	    }
	}
	if err := rows.Err(); err != nil {
	    log.Println(err)
	}
	return err
}

func (h *IPLimit) LogAttempt(db *sqlx.DB) (err error) {
	h.LastLoginAttemptDate.Time = time.Now()
	h.LastLoginAttemptDate.Valid = true

	stmt, err := db.Prepare("INSERT INTO `IPLimit` (`IP`, `LastLoginAttemptDate`, `UsersAllowedCount`) " +
		"VALUES (?, ?, ?) " +
		"ON DUPLICATE KEY UPDATE `LastLoginAttemptDate` = VALUES(`LastLoginAttemptDate`)")
	if err != nil {
	    log.Println(err)
	    return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(h.IP, h.LastLoginAttemptDate, NewUsersPerIPPerDay)
	if err != nil {
	    log.Println(err)
	    return err
	}
	return nil
}

func (h *IPLimit) LogNewUser(db *sqlx.DB) (err error) {
	if !h.CountResetDate.Valid || h.CountResetDate.Time.Before(time.Now()) {
		h.CountResetDate.Time = time.Now().Add(24 * time.Hour)
		h.CountResetDate.Valid = true
		h.UsersAllowedCount = NewUsersPerIPPerDay - 1
	} else {
		h.UsersAllowedCount -= 1
	}

	stmt, err := db.Prepare("INSERT INTO `IPLimit` (`IP`, `UsersAllowedCount`, `CountResetDate`) " +
		"VALUES (?, ?, ?) " +
		"ON DUPLICATE KEY UPDATE `UsersAllowedCount` = VALUES(`UsersAllowedCount`), `CountResetDate` = VALUES(`CountResetDate`)")
	if err != nil {
	    log.Println(err)
	    return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(h.IP, h.UsersAllowedCount, h.CountResetDate)
	if err != nil {
	    log.Println(err)
	    return err
	}
	return nil
}

func (h *IPLimit) Clear(db *sqlx.DB) (err error) {
	stmt, err := db.Prepare("DELETE FROM `IPLimit` WHERE IP LIKE ?")
	if err != nil {
	    log.Println(err)
	    return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(h.IP)
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
	return nil
}

