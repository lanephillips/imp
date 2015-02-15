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

    // Heroku uses env var to specify port
    port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}

	// any requests on the regular HTTP port get automatically redirected to the secure home page
	// don't just redirect HTTP to HTTPS, because that doesn't train the user to not use HTTP on the first request
	// TODO: commented out so I can run Apache on 80 on my computer
	// TODO: this didn't work, maybe I have too many web servers running, also Chrome gets pretty confused with port numbers
	// go http.ListenAndServe(cfg.Server.Host + ":80", http.HandlerFunc(func (w http.ResponseWriter, req *http.Request) {
	// 	// TODO: secure app shouldn't have to be at the root of the domain
	// 	http.Redirect(w, req, "https://" + cfg.Server.Host + ":" + port, http.StatusMovedPermanently)
	// }))

    r := mux.NewRouter()
    r.HandleFunc("/", HomeHandler)

    api := r.PathPrefix("/api").Subrouter()
    api.HandleFunc("/user/{id}", UsersHandler)

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
