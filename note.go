package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sort"
)

const MaximumNotesReturned = 100

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

// TODO: use Marshaler interface
func (n *Note) AsMap() *map[string]interface{} {
	m := map[string]interface{}{
		"NoteId": n.NoteId,
		"UserId": n.UserId,
		"Text": n.Text,
		"Date": n.Date.Time.Unix(),
		"Edited": n.Edited,
		"GroupId": n.GroupId,
	}
	if n.Link.Valid {
		m["Link"] = n.Link.String
	}
	if n.LinkType.Valid {
		m["LinkType"] = n.LinkType.String
	}
	return &m
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
		dagrx := regexp.MustCompile("‡(\\d+)")
		daggers := dagrx.FindAllStringSubmatch(note.Text, -1)
		if daggers != nil {
			for _, dagger := range daggers {
				d2, _ := strconv.Atoi(dagger[1])
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

// first call r.ParseForm()
func validIntFormValue(r *http.Request, fieldName string, defaultValue int) int {
	stringVal := r.PostFormValue(fieldName)
	if len(stringVal) == 0 {
		return defaultValue
	}
	intVal, err := strconv.Atoi(stringVal)
	if err != nil {
		// TODO: or should we 400 on malformed ints?
		return defaultValue
	}
	return intVal
}

func ListNotesHandler(rw http.ResponseWriter, r *http.Request) {
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

	// TODO: guest authorization and user id parm

	r.ParseForm()
	where := " WHERE UserId = ?"
	sinceId := validIntFormValue(r, "since_id", 0)
	if sinceId > 0 {
		where += " AND NoteId > " + strconv.Itoa(sinceId)
	}
	sinceDate := validIntFormValue(r, "since_date", 0)
	if sinceDate > 0 {
		where += " AND Date > FROM_UNIXTIME(" + strconv.Itoa(sinceDate) + ")"
	}
	beforeId := validIntFormValue(r, "before_id", 0)
	if beforeId > 0 {
		where += " AND NoteId < " + strconv.Itoa(beforeId)
	}
	beforeDate := validIntFormValue(r, "before_date", 0)
	if beforeDate > 0 {
		where += " AND Date < FROM_UNIXTIME(" + strconv.Itoa(beforeDate) + ")"
	}

	count := validIntFormValue(r, "count", MaximumNotesReturned)
	if count > MaximumNotesReturned {
		count = MaximumNotesReturned
	}
	limit := " LIMIT " + strconv.Itoa(count)

	notes := []Note{}
	err = db.Select(&notes, "SELECT * FROM Note" + where + limit, token.UserId)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	notes2 := []interface{}{}
	for _, note := range notes {
		notes2 = append(notes2, note.AsMap())
	}

	sendData(rw, http.StatusOK, notes2)
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

	sendData(rw, http.StatusCreated, note.AsMap())
}

func GetNoteHandler(rw http.ResponseWriter, r *http.Request) {
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

	noteId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: guest authorization

	note := new(Note)
	err = db.Get(note, "SELECT * FROM Note WHERE NoteId = ?", noteId)
	if err == sql.ErrNoRows {
		sendError(rw, http.StatusNotFound, "There is no note with that ID.")
		return
	} else if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	sendData(rw, http.StatusOK, note.AsMap())
}

func PutNoteHandler(rw http.ResponseWriter, r *http.Request) {
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

	noteId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusBadRequest, err.Error())
		return
	}

	note := new(Note)
	err = db.Get(note, "SELECT * FROM Note WHERE NoteId = ?", noteId)
	if err == sql.ErrNoRows {
		sendError(rw, http.StatusNotFound, "There is no note with that ID.")
		return
	} else if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	} else if note.UserId != token.UserId {
		sendError(rw, http.StatusUnauthorized, "Only the note's author may edit it.")
		return
	}

	r.ParseForm()
	noteText := r.PostFormValue("note")
	note2 := parseNote(noteText)
	if note2 == nil {
		sendError(rw, http.StatusBadRequest, "Bad Request")
		return
	}
	note.Text = note2.Text
	note.Link = note2.Link
	note.Edited = true

	_, err = db.NamedExec("UPDATE Note SET Text = :Text, Link = :Link, Edited = 1 "+
		"WHERE NoteId = :NoteId", note)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	sendData(rw, http.StatusOK, note.AsMap())
}

func DeleteNoteHandler(rw http.ResponseWriter, r *http.Request) {
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

	// TODO: noteId parm middleware
	noteId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: owner auth middleware
	var noteUser int64
	err = db.Get(&noteUser, "SELECT UserId FROM Note WHERE NoteId = ?", noteId)
	if err == sql.ErrNoRows {
		sendError(rw, http.StatusNotFound, "There is no note with that ID.")
		return
	} else if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	} else if noteUser != token.UserId {
		sendError(rw, http.StatusUnauthorized, "Only the note's author may delete it.")
		return
	}

	_, err = db.Exec("DELETE FROM `Note` WHERE NoteId = ?", noteId)
	if err != nil {
		fmt.Println(err)
		sendError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	sendData(rw, http.StatusNoContent, "")
}

