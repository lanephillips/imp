package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	IMPLocationHeader = "IMP-API-Location"
	IMPDefaultPort = "5039"		// 443 would actually be our first choice
)

var cfg Config
var db *sqlx.DB

func sendError(rw http.ResponseWriter, status int, message string) {
	rw.Header().Set(IMPLocationHeader, cfg.Api.Version + ";" + cfg.Api.Location)
	envelope := map[string]interface{}{
		"errors": [1]interface{}{
			map[string]interface{}{
				"status": fmt.Sprintf("%d",status),
				"title": message,
				"detail": message,
			},
		},
	}
	render.New().JSON(rw, status, envelope)
}

func sendData(rw http.ResponseWriter, status int, data interface{}) {
	rw.Header().Set(IMPLocationHeader, cfg.Api.Version + ";" + cfg.Api.Location)
	envelope := map[string]interface{}{
		"data": data,
	}
	render.New().JSON(rw, status, envelope)
}

func getIP(r *http.Request) string {
    if ipProxy := r.Header.Get("X-Forwarded-For"); len(ipProxy) > 0 {
    	ips := strings.Split(ipProxy, ", ")
        return ips[0]
    }
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}

func main() {
	err := LoadConfigInto(&cfg, "config.gcfg")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Loaded config.")

	// set up database connection
	db, err = sqlx.Open("mysql", cfg.Database.User + ":" + cfg.Database.Password + "@/" + cfg.Database.Database)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Opened database.")

	// DB fields are capitalized in the same way as Go structs, so mapper is a no-op
	db.MapperFunc(func(s string) string {
		return s
	})

	// set up routes
	r := mux.NewRouter()
    r.HandleFunc("/",  func (rw http.ResponseWriter, r *http.Request) {
    	sendData(rw, http.StatusOK, "")
	})

	// authentication
    r.HandleFunc("/token", PostTokenHandler).Methods("POST")
    r.HandleFunc("/token/{token}", DeleteTokenHandler).Methods("DELETE")

    // guest authentication
    r.HandleFunc("/user/{handle}/host/{host}", GetUserHostHandler).Methods("GET")
    r.HandleFunc("/user/{handle}/host", PostUserHostHandler).Methods("POST")
    r.HandleFunc("/guest", PostGuestHandler).Methods("POST")

    // users
	r.HandleFunc("/user", PostUserHandler).Methods("POST")

	// notes
	r.HandleFunc("/note", ListNotesHandler).Methods("GET")
	r.HandleFunc("/note", PostNoteHandler).Methods("POST")
	r.HandleFunc("/note/{id}", GetNoteHandler).Methods("GET")
	r.HandleFunc("/note/{id}", PutNoteHandler).Methods("PUT")
	r.HandleFunc("/note/{id}", DeleteNoteHandler).Methods("DELETE")

	// groups
	r.HandleFunc("/group", NotImplementedHandler).Methods("GET")
	r.HandleFunc("/group", NotImplementedHandler).Methods("POST")
	r.HandleFunc("/group/{id}", NotImplementedHandler).Methods("GET")
	r.HandleFunc("/group/{id}", NotImplementedHandler).Methods("PUT")
	r.HandleFunc("/group/{id}", NotImplementedHandler).Methods("DELETE")
	r.HandleFunc("/group/{id}/{address}", NotImplementedHandler).Methods("PUT")
	r.HandleFunc("/group/{id}/{address}", NotImplementedHandler).Methods("DELETE")

	// mutes and blocks
	r.HandleFunc("/user/{handle}/mute", NotImplementedHandler).Methods("GET")
	r.HandleFunc("/user/{handle}/mute/{address}", NotImplementedHandler).Methods("PUT")
	r.HandleFunc("/user/{handle}/mute/{address}", NotImplementedHandler).Methods("DELETE")
	r.HandleFunc("/user/{handle}/block", NotImplementedHandler).Methods("GET")
	r.HandleFunc("/user/{handle}/block/{address}", NotImplementedHandler).Methods("PUT")
	r.HandleFunc("/user/{handle}/block/{address}", NotImplementedHandler).Methods("DELETE")

    // Heroku uses env var to specify port
    port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}
	if port == "" {
		port = IMPDefaultPort
	}

    hostname := cfg.Server.Host + ":" + port
    log.Println("Listening on " + hostname + ".")
	http.ListenAndServeTLS(hostname, cfg.Server.Certificate, cfg.Server.Key, r)
}

func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	sendError(rw, http.StatusNotImplemented, "Not Implemented")
}

