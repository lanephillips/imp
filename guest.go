package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// allow at least this much time for the token exchange transaction to complete
const GuestTokenTimeout = 30

type Host struct {
	HostId int64
	Name string
	Location string
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
	// TODO: auth middleware
	token, err := FetchToken(db, r)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	if token == nil {
		fmt.Println(err)
		sendError(rw, http.StatusUnauthorized, "Unauthorized")
		return
	}

	handle := mux.Vars(r)["handle"]	
	hostname := mux.Vars(r)["host"]

	var user User
	err = db.Get(&user, "SELECT UserId, Handle FROM User WHERE Handle LIKE ?", handle)
	if err == sql.ErrNoRows {
		sendError(rw, http.StatusNotFound, err.Error())
		return
	} else if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	if token.UserId != user.UserId {
		fmt.Println(err)
		sendError(rw, http.StatusUnauthorized, "Unauthorized")
		return
	}

	host, err := FetchHost(db, hostname)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	var userHost UserHost
	err = db.Get(&userHost, "SELECT * FROM UserHost WHERE UserId = ? AND HostId = ?", user.UserId, host.HostId)
	if err == nil && len(userHost.Token) > 0 {
		// we found it
		sendData(rw, http.StatusOK, map[string]interface{}{
				"host": hostname,
				"token": userHost.Token,
			})
		return
	} else if err == nil && time.Now().Before(userHost.CreatedDate.Time.Add(time.Duration(GuestTokenTimeout) * time.Second)) {
		// there is already a request pending
		sendError(rw, 429, "Too many requests for this host.")
		return
	} else if err != nil && err != sql.ErrNoRows {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	userHost.UserId = user.UserId
	userHost.HostId = host.HostId
	userHost.Nonce = RandomString(50)
	userHost.Token = ""
	userHost.CreatedDate.Time = time.Now()
	userHost.CreatedDate.Valid = true

	_, err = db.NamedExec("INSERT INTO `UserHost` (`UserId`, `HostId`, `Nonce`, `Token`, `CreatedDate`) " +
				"VALUES (:UserId, :HostId, :Nonce, :Token, :CreatedDate) " +
				"ON DUPLICATE KEY UPDATE `Nonce` = :Nonce, `Token` = :Token, `CreatedDate` = :CreatedDate", &userHost)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	// at this point we don't know how the foreign host will respond, so send 202 Accepted
	sendData(rw, http.StatusAccepted, "")

	go func() {
		if len(host.Location) == 0 {
			err := DiscoverHost(db, host)
			if err != nil {
				log.Println(err)
				return
			}
		}

		resp, err := http.PostForm("https://" + host.Location + "/guest",
			url.Values{"host": {cfg.Api.Host}, "handle": {user.Handle}, "nonce": {userHost.Nonce}})
		if err != nil {
		    log.Println(err)
		    return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
	        log.Println(resp.Status)
			bodyBytes, _ := ioutil.ReadAll(resp.Body) 
	        log.Println(string(bodyBytes))
		}
	}()
}

// called by foreign host to place an access token for user of this host
func PostUserHostHandler(rw http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	hostname := r.PostFormValue("host")
	if len(hostname) == 0 {
		sendError(rw, http.StatusBadRequest, "Host is missing.")
		return
	}
	token := r.PostFormValue("token")
	if len(token) == 0 {
		sendError(rw, http.StatusBadRequest, "Token is missing.")
		return
	}
	nonce := r.PostFormValue("nonce")
	if len(nonce) == 0 {
		sendError(rw, http.StatusBadRequest, "Nonce is missing.")
		return
	}

	handle := mux.Vars(r)["handle"]	

	var user User
	err := db.Get(&user, "SELECT UserId, Handle FROM User WHERE Handle LIKE ?", handle)
	if err == sql.ErrNoRows {
		sendError(rw, http.StatusNotFound, "There is no user with that handle.")
		return
	} else if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	host, err := FetchHost(db, hostname)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	var userHost UserHost
	err = db.Get(&userHost, "SELECT * FROM UserHost WHERE UserId = ? AND HostId = ? AND Nonce LIKE ?",
		user.UserId, host.HostId, nonce)
	if err == sql.ErrNoRows {
		sendError(rw, http.StatusUnauthorized, "The user did not request a guest token.")
		return
	} else if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = db.NamedExec("UPDATE UserHost SET Nonce = '', Token = :Token WHERE UserId = :UserId AND HostId = :HostId", &userHost)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	sendData(rw, http.StatusOK, "")
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

	host, err := FetchHost(db, hostname)
	if err != nil {
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
		if len(host.Location) == 0 {
			err := DiscoverHost(db, host)
			if err != nil {
				log.Println(err)
				return
			}
		}

		resp, err := http.PostForm("https://" + host.Location + "/user/" + guest.Handle + "/host",
			url.Values{"host": {cfg.Api.Host}, "token": {guest.Token}, "nonce": {nonce}})
		if err != nil {
		    log.Println(err)
		    return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
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

func FetchHost(db *sqlx.DB, hostname string)  (*Host, error) {
	var host Host
	err := db.Get(&host, "SELECT * FROM Host WHERE Name LIKE ?", hostname)
	host.Name = hostname

	if err == sql.ErrNoRows {
		result, err := db.Exec("INSERT INTO `Host` (`Name`) VALUES (?)", hostname)
		if err != nil {
			return nil, err
		}
		host.HostId, err = result.LastInsertId()
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &host, nil
}

func DiscoverHost(db *sqlx.DB, host *Host) error {
	urls := []string{
		"https://" + host.Name,
		"https://" + host.Name + ":" + IMPDefaultPort,
		"http://" + host.Name,
	}

	for len(urls) > 0 {
		lurl, urls := urls[0], urls[1:]

		// this will follow redirects
		resp, err := http.Get(lurl)
		if err != nil || resp.StatusCode != http.StatusOK {
			// that didn't work, try the next one
			continue
		}
		defer resp.Body.Close()

		location := resp.Header.Get(IMPLocationHeader)
		if len(location) > 0 {
			s := strings.Split(location, ";")
			location = s[len(s) - 1]

			if lurl == "https://" + location {
				// we found it! the API location we got is the URL we're looking at
				host.Location = location

				_, err = db.Exec("UPDATE Host SET Location = ? WHERE HostId = ?", host.Location, host.HostId)
				if err != nil {
					log.Println(err)
				}
				return nil
			}

			// it's good that we found our header, but let's go there to verify
			// put it in the front of the queue
			// TODO: server misconfiguration could lead to infinite loop?
			urls = append([]string{ "https://" + location }, urls...)
			continue
		}

		// TODO: parse html with golang.org/x/net/html 
		// TODO: and look for <meta http-equiv=...
	}
	return errors.New("Could not locate the IMP host.")
}
