package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
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

	// set up database connection
	db, err := sql.Open("mysql", cfg.Database.User + ":" + cfg.Database.Password + "@/" + cfg.Database.Database)
	if err != nil {
	    panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
	    panic(err.Error()) // proper error handling instead of panic in your app
	}

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

    api.HandleFunc("/user", func (rw http.ResponseWriter, r *http.Request) {
    	// TODO: validate and create user
	}).Methods("POST")

    api.HandleFunc("/user/{handle}", func (rw http.ResponseWriter, r *http.Request) {
	    handle := mux.Vars(r)["handle"]

	    var u User
	    u.Fetch(db, handle)

		rend := render.New()
	    rend.JSON(rw, http.StatusOK, u)

	    fmt.Fprintln(rw, "showing user", handle)
	}).Methods("GET")

    api.HandleFunc("/user/{handle}", func (rw http.ResponseWriter, r *http.Request) {
    	// TODO: update user
	    handle := mux.Vars(r)["handle"]

	    var u User
	    u.Fetch(db, handle)

		rend := render.New()
	    rend.JSON(rw, http.StatusOK, u)

	    fmt.Fprintln(rw, "showing user", handle)
	}).Methods("PUT")

    api.HandleFunc("/user/{handle}", func (rw http.ResponseWriter, r *http.Request) {
    	// TODO: delete user
	    handle := mux.Vars(r)["handle"]

	    var u User
	    u.Fetch(db, handle)

		rend := render.New()
	    rend.JSON(rw, http.StatusOK, u)

	    fmt.Fprintln(rw, "showing user", handle)
	}).Methods("DELETE")

	http.ListenAndServeTLS(cfg.Server.Host + ":" + port, cfg.Server.Certificate, cfg.Server.Key, r)
}

func HomeHandler(rw http.ResponseWriter, r *http.Request) {
	rend := render.New()
	rend.HTML(rw, http.StatusOK, "login", nil)
}
