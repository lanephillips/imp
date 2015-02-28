package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sort"
)

type Note struct {
	NoteId int64
	UserId int64
	Text string
	Link sql.NullString
	LinkType sql.NullString
	Date mysql.NullTime
	Edited bool
	Deleted bool
	GroupId int64
}

// interface for sorting strings by length
type ByLength []string
func (s ByLength) Len() int {
    return len(s)
}
func (s ByLength) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByLength) Less(i, j int) bool {
    return len(s[i]) < len(s[j])
}

func parseNote(text string) (note *Note) {
	note = new(Note)
	note.Text = text

	// TODO: find @mentions
	// atrx := MustCompile("@([[:alnum:]]{1,16})@((?:(?:(?:[a-zA-Z0-9][-a-zA-Z0-9]*)?[a-zA-Z0-9])[.])*(?:[a-zA-Z][-a-zA-Z0-9]*[a-zA-Z0-9]|[a-zA-Z]))")
	// TODO: to be shortened, @mentions must be users that exist and are not blocking this user
	// TODO: defer that process so we can return to the user right away and return 202 Accepted

	// TODO: we will replace @mentions with @<number>
	// TODO: find the highest @<number> that doesn't collide with existing text
	// clients will substitute mentions for highest @<numbers> in reverse order

	// find all things that look like links
	linkrx := regexp.MustCompile("\\b(?:https?|ftp)://\\S+")
	matches := linkrx.FindAllString(note.Text, 0)

	if matches != nil {
		// find the longest link
		sort.Sort(ByLength(matches))
		note.Link.String = matches[len(matches) - 1]
		note.Link.Valid = true

		// ‡<number> indicates location to insert link
		dagIdx := 0
		// if for some strange reason someone actually typed that, we search for the highest non-colliding index
		dagrx := regexp.MustCompile("‡\\d+")
		daggers := dagrx.FindAllString(note.Text, 0)
		if daggers != nil {
			for _, dagger := range daggers {
				d2, _ := strconv.Atoi(dagger[1:])
				if d2 >= dagIdx {
					dagIdx = d2 + 1
				}
			}
		}

		// replace the link with our symbol
		// clients can later replace ‡<highest number> with the link
		dagger := fmt.Sprintf("‡%d", dagIdx)
		note.Text = strings.Replace(note.Text, note.Link.String, dagger, 1)
	}

	if len(note.Text) > 140 {
		// TODO: return error?
		return nil
	}
	return note
}

func ListNotesHandler(rw http.ResponseWriter, r *http.Request) {
}

func PostNoteHandler(rw http.ResponseWriter, r *http.Request) {
	// ip := getIP(r)
	// // fmt.Println("client ip is", ip)

	// // rate limit new user creation by ip
	// ipLimit, err := FetchIPLimit(db, ip)
	// if err != nil {
	// 	fmt.Println(err)
	// 	sendError(rw, http.StatusInternalServerError, err.Error())
	// 	return
	// }
	// if ipLimit.UsersAllowedCount <= 0 && ipLimit.CountResetDate.Valid && time.Now().Before(ipLimit.CountResetDate.Time) {
	// 	fmt.Println("Too many new user accounts from this address.", ipLimit)
	// 	sendError(rw, 429, "Too many new user accounts from this address.")
	// 	return
	// }

	// r.ParseForm()

	// handle := r.PostFormValue("handle")
	// if len(handle) == 0 {
	// 	fmt.Println("Missing handle.")
 //    	sendError(rw, http.StatusBadRequest, "Missing handle.")
	// 	return
	// }

	// email, err := mail.ParseAddress(r.PostFormValue("email"))
	// if err != nil {
	// 	fmt.Println(err)
 //    	sendError(rw, http.StatusBadRequest, "Invalid email address.")
	// 	return
	// }

 //    // TODO: check password strength
	// password := r.PostFormValue("password")
	// if len(password) == 0 {
	// 	fmt.Println("Missing password.")
 //    	sendError(rw, http.StatusBadRequest, "Missing password.")
	// 	return
	// }

	// // look up handle to see if this user already exists
 //    var count int64
 //    err = db.Get(&count, "SELECT COUNT(*) FROM User WHERE Handle LIKE ?", handle)
	// if err != nil {
	// 	fmt.Println(err)
	// 	sendError(rw, http.StatusInternalServerError, err.Error())
	// 	return
	// }
 //    if count > 0 {
	// 	fmt.Println("Handle already in use.")
 //    	sendError(rw, http.StatusConflict, "That handle is already in use.")
	// 	return
 //    }

 //    err = db.Get(&count, "SELECT COUNT(*) FROM User WHERE Email LIKE ?", email.Address)
	// if err != nil {
	// 	fmt.Println(err)
	// 	sendError(rw, http.StatusInternalServerError, err.Error())
	// 	return
	// }
 //    if count > 0 {
	// 	fmt.Println("Email already in use.")
 //    	sendError(rw, http.StatusConflict, "That email address is already in use.")
	// 	return
 //    }

 //    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// if err != nil {
	// 	fmt.Println(err)
	// 	sendError(rw, http.StatusInternalServerError, err.Error())
	// 	return
	// }

	// var u User
 //    u.Handle = handle
 //    u.Email = email.Address
 //    u.Status = ""
 //    u.Biography = ""
 //    u.PasswordHash = string(hash)
	// fmt.Println(u)

	// result, err := db.NamedExec("INSERT INTO `User` (`Handle`, `Status`, `Biography`, `Email`, `PasswordHash`) " +
	// 		"VALUES (:Handle, :Status, :Biography, :Email, :PasswordHash)", &u)
	// if err != nil {
	// 	fmt.Println(err)
	// 	sendError(rw, http.StatusInternalServerError, err.Error())
	// 	return
	// }
	// u.UserId, err = result.LastInsertId()
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// // go ahead and log user in
	// t, err := MakeToken(db, &u)
	// if err != nil {
	// 	fmt.Println(err)
	// 	// something went wrong, but at least we created the user, so don't die here
	// }
	// ipLimit.LogNewUser(db)

	// resp := map[string]interface{}{
	// 	"user": &u,
	// 	"token": t.Token,
	// }

	// sendData(rw, http.StatusCreated, resp)
}

func GetNoteHandler(rw http.ResponseWriter, r *http.Request) {
}

func PutNoteHandler(rw http.ResponseWriter, r *http.Request) {
}

func DeleteNoteHandler(rw http.ResponseWriter, r *http.Request) {
}

