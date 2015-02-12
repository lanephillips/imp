package main

import (
	"fmt"
	"net/http"
	"os"
	"code.google.com/p/gcfg"
	"github.com/gorilla/mux"
	// "database/sql"
	// _ "github.com/go-sql-driver/mysql"
	// "log"
	// "github.com/unrolled/render"
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
    id := mux.Vars(r)["id"]
    fmt.Fprintln(rw, "showing user", id)
}
