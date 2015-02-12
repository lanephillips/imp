package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"github.com/unrolled/render"
)

var cfg Config

func main() {
	err := LoadConfigInto(&cfg, "config.gcfg")
	if err != nil {
		fmt.Println(err)
		// TODO: die
	}
	// fmt.Println("loaded config", cfg)

    r := mux.NewRouter()
    r.HandleFunc("/", HomeHandler)

    api := r.PathPrefix("/api").Subrouter()
    api.HandleFunc("/user/{id}", UsersHandler)

    // Heroku uses this to specify port
    port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}
	// TODO: permanently redirect non-SSL to SSL, Chrome actually downloads 7 bytes of something if I use http:
	// TODO: keyfile locations should be in config file
	// TODO: we probably want to use this: http://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security
	http.ListenAndServeTLS(cfg.Server.Host + ":" + port, cfg.Server.Certificate, cfg.Server.Key, r)
}

func HomeHandler(rw http.ResponseWriter, r *http.Request) {
	rend := render.New()
	rend.HTML(rw, http.StatusOK, "login", nil)
}

func UsersHandler(rw http.ResponseWriter, r *http.Request) {
	dsn := fmt.Sprint(cfg.Database.User, ":", cfg.Database.Password, "@/", cfg.Database.Database)
	db, err := sql.Open("mysql", dsn)
	// fmt.Println("%v %v", db, err)

    id := mux.Vars(r)["id"]

	rows, err := db.Query("SELECT `UserId`, `Handle`, `Status`, `Biography`, `JoinedDate` FROM `User` WHERE 1")	// TODO: user id
	if err != nil {
	    log.Fatal(err)
	    // TODO: 500
	}

	var users []*User
	for rows.Next() {
	    var user User
	    if err := rows.Scan(&user.UserId, &user.Handle, &user.Status, &user.Biography, &user.JoinedDate); err != nil {
	        log.Fatal(err)
	    }
	    users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
	    log.Fatal(err)
	    // TODO: 500
	}

	rend := render.New()
    rend.JSON(rw, http.StatusOK, users)

    fmt.Fprintln(rw, "showing user", id)
}
