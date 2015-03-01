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
	linkrx := regexp.MustCompile("\\b(?i:https?|ftp)://\\S+")
	matches := linkrx.FindAllString(note.Text, -1)
	//fmt.Println(matches)

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

	if len(note.Text) > 140 || len(note.Text) == 0 {
		// TODO: return error?
		return nil
	}
	return note
}

func ListNotesHandler(rw http.ResponseWriter, r *http.Request) {
}

func PostNoteHandler(rw http.ResponseWriter, r *http.Request) {
	// TODO: we should make an auth middleware, once I can wrap my head around that
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

	r.ParseForm()

	noteText := r.PostFormValue("note")
	note := parseNote(noteText)

	if note == nil {
		sendError(rw, http.StatusBadRequest, "Bad Request")
		return
	}

	// TODO: groups
	group := r.PostFormValue("group")
	if len(group) > 0 {
		groupId, _ := strconv.Atoi(group)
		note.GroupId = int64(groupId)
	}

	// TODO: defer processing mentions

	note.UserId = token.UserId
	result, err := db.NamedExec("INSERT INTO `Note` (`UserId`, `Text`, `Link`, `LinkType`, `GroupId`) " +
			"VALUES (:UserId, :Text, :Link, :LinkType, :GroupId)", note)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	note.NoteId, err = result.LastInsertId()
	if err != nil {
		fmt.Println(err)
	}

	sendData(rw, http.StatusCreated, note)
}

func GetNoteHandler(rw http.ResponseWriter, r *http.Request) {
}

func PutNoteHandler(rw http.ResponseWriter, r *http.Request) {
}

func DeleteNoteHandler(rw http.ResponseWriter, r *http.Request) {
}

