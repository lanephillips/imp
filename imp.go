package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/unrolled/render"
	"golang.org/x/crypto/bcrypt"
)

var cfg Config

func sendError(rw http.ResponseWriter, status int, message string) {
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
	envelope := map[string]interface{}{
		"data": data,
	}
	render.New().JSON(rw, status, envelope)
}

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
    	r.ParseForm()

    	handle := r.PostFormValue("handle")
    	if len(handle) == 0 {
			fmt.Println("Missing handle.")
	    	sendError(rw, http.StatusBadRequest, "Missing handle.")
			return
    	}

	    // TODO: validate email
    	email := r.PostFormValue("email")
    	if len(email) == 0 {
			fmt.Println("Missing email.")
	    	sendError(rw, http.StatusBadRequest, "Missing email.")
			return
    	}

	    // TODO: check password strength
    	password := r.PostFormValue("password")
    	if len(password) == 0 {
			fmt.Println("Missing password.")
	    	sendError(rw, http.StatusBadRequest, "Missing password.")
			return
    	}

    	// look up handle to see if this user already exists
	    var u User
	    err = u.Fetch(db, handle)
		if err != nil {
			fmt.Println(err)
			sendError(rw, http.StatusInternalServerError, err.Error())
			return
		}

	    if u.UserId > 0 {
			fmt.Println("Handle already in use.")
	    	// TODO: we need to define a regular envelope format
	    	sendError(rw, http.StatusConflict, "That handle is already in use.")
			return
	    }

	    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println(err)
			sendError(rw, http.StatusInternalServerError, err.Error())
			return
		}

	    u.UserId = -1
	    u.Handle = handle
	    u.Email = email
	    u.PasswordHash = string(hash)
		fmt.Println(u)

	    err = u.Save(db)
		if err != nil {
			fmt.Println(err)
			sendError(rw, http.StatusInternalServerError, err.Error())
			return
		}

		// TODO: auth
		sendData(rw, http.StatusCreated, u)
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
