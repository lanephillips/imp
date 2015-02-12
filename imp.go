package main

import (
	"fmt"
	"net/http"
	"os"
	"code.google.com/p/gcfg"
	"github.com/gorilla/mux"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"github.com/unrolled/render"
)

type Config struct {
	Database struct {
		Database string
		User string
		Password string
	}
}

var cfg Config

func main() {
	err := gcfg.ReadFileInto(&cfg, "config.gcfg")
	if err != nil {
		fmt.Println(err)
		// TODO: 500
	}
	//fmt.Println("loaded config", cfg)

    r := mux.NewRouter()
    r.HandleFunc("/", HomeHandler)

    api := r.PathPrefix("/api").Subrouter()
    api.HandleFunc("/user/{id}", UsersHandler)

    port := os.Getenv("PORT")
	if port == "" {
	  port = "8080"
	}
	http.ListenAndServe(":"+port, r)

}

func HomeHandler(rw http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(rw, "Home")
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
