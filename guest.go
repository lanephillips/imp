package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// allow at least this much time for the token exchange transaction to complete
const GuestTokenTimeout = 30

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
	Nonce string
	Token string
	CreatedDate mysql.NullTime
}

// called by user of this host to get token for accessing foreign host
func GetUserHostHandler(rw http.ResponseWriter, r *http.Request) {
	
}

// called by foreign host to place an access token for user of this host
func PostUserHostHandler(rw http.ResponseWriter, r *http.Request) {
	
}

// called by foreign host to request access token for one of its users
func PostGuestHandler(rw http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	handle := r.PostFormValue("handle")
	if len(handle) == 0 {
		sendError(rw, http.StatusBadRequest, "Handle is missing.")
		return
	}
	hostname := r.PostFormValue("host")
	if len(hostname) == 0 {
		sendError(rw, http.StatusBadRequest, "Host is missing.")
		return
	}
	nonce := r.PostFormValue("nonce")
	if len(nonce) == 0 {
		sendError(rw, http.StatusBadRequest, "Nonce is missing.")
		return
	}

	var host Host
	err := db.Get(&host, "SELECT HostId FROM Host WHERE Name LIKE ?", hostname)
	host.Name = hostname

	if err == sql.ErrNoRows {
		result, err := db.Exec("INSERT INTO `Host` (`Name`) VALUES (?)", hostname)
		if err != nil {
			fmt.Println(err)
			sendError(rw, http.StatusInternalServerError, err.Error())
			return
		}
		host.HostId, err = result.LastInsertId()
		if err != nil {
			fmt.Println(err)
			sendError(rw, http.StatusInternalServerError, err.Error())
			return
		}
	} else if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	var guest Guest
	err = db.Get(&guest, "SELECT * FROM Guest WHERE Handle LIKE ? AND HostId = ?", handle, host.HostId)
	if err == sql.ErrNoRows {
		guest.Handle = handle
		guest.HostId = host.HostId
		guest.CreatedDate.Time = time.Now()
		guest.CreatedDate.Valid = true
	} else if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	} else {
		// already exists, rate limit requests
		if time.Now().Before(guest.CreatedDate.Time.Add(time.Duration(GuestTokenTimeout) * time.Second)) {
			sendError(rw, 429, "Too many requests for this guest.")
			return
		}
	}

	guest.Token = RandomString(50)

	// at this point we don't know how the user's host will respond, so send 202 Accepted
	sendData(rw, http.StatusAccepted, "")

	go func() {
		// TODO: api prefix
		resp, err := http.PostForm("https://" + host.Name + "/user/" + guest.Handle + "/host",
			url.Values{"host": {cfg.Server.Host}, "token": {guest.Token}, "nonce": {nonce}})
		if err != nil {
		    log.Println(err)
		    return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// don't save until we get a successful response,
			// otherwise an attacker could destroy guest tokens by posting this request with a bum nonce
			_, err = db.NamedExec("INSERT INTO `Guest` (`Handle`, `HostId`, `Token`, `CreatedDate`) " +
				"VALUES (:Handle, :HostId, :Token, :CreatedDate) " +
				"ON DUPLICATE KEY UPDATE `Token` = VALUES(`Token`)", &guest)
			if err != nil {
			    log.Println(err)
			    return
			}
		} else {
	        log.Println(resp.Status)
			bodyBytes, _ := ioutil.ReadAll(resp.Body) 
	        log.Println(string(bodyBytes))
		}
	}()
}
