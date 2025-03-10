package manage

import (
	"net/http"

	"github.com/gochan-org/gochan/pkg/gcsql"
	"github.com/rs/zerolog"
)

const (
	// NoPerms allows anyone to access this Action
	NoPerms = iota
	// JanitorPerms allows anyone with at least a janitor-level account to access this Action
	JanitorPerms
	// ModPerms allows anyone with at least a moderator-level account to access this Action
	ModPerms
	// AdminPerms allows only the site administrator to view this Action
	AdminPerms
)

const (
	// NoJSON actions will return an error if JSON is requested by the user
	NoJSON = iota
	// OptionalJSON actions have an optional JSON output if requested
	OptionalJSON
	// AlwaysJSON actions always return JSON whether or not it is requested
	AlwaysJSON
)

type CallbackFunction func(writer http.ResponseWriter, request *http.Request, staff *gcsql.Staff, wantsJSON bool, infoEv *zerolog.Event, errEv *zerolog.Event) (output interface{}, err error)

// Action represents the functions accessed by staff members at /manage/<functionname>.
type Action struct {
	// the string used when the user requests /manage/<ID>
	ID string `json:"id"`

	// The text shown in the staff menu and the window title
	Title string `json:"title"`

	// Permissions represent who can access the page. 0 for anyone,
	// 1 requires the user to have a janitor, mod, or admin account. 2 requires mod or admin,
	// and 3 is only accessible by admins
	Permissions int `json:"perms"`

	// JSONoutput sets what the action can output. If it is 0, it will throw an error if
	// JSON is requested. If it is 1, it can output JSON if requested, and if 2, it always
	// outputs JSON whether it is requested or not
	JSONoutput int `json:"jsonOutput"` // if it can sometimes return JSON, this should still be false

	// Callback executes the staff page. if wantsJSON is true, it should return an object
	// to be marshalled into JSON. Otherwise, a string assumed to be valid HTML is returned.
	//
	// IMPORTANT: the writer parameter should only be written to if absolutely necessary (for example,
	// if a redirect wouldn't work in handler.go) and even then, it should be done sparingly
	Callback CallbackFunction `json:"-"`
}

var actions []Action

// returns the action by its ID, or nil if it doesn't exist
func getAction(id string, rank int) *Action {
	for a := range actions {
		if rank == NoPerms && actions[a].Permissions > NoPerms {
			id = "login"
		}
		if actions[a].ID == id {
			return &actions[a]
		}
	}
	return nil
}

func RegisterManagePage(id string, title string, permissions int, jsonOutput int, callback CallbackFunction) {
	actions = append(actions, Action{
		ID:          id,
		Title:       title,
		Permissions: permissions,
		JSONoutput:  jsonOutput,
		Callback:    callback,
	})
}

func getAvailableActions(rank int, noJSON bool) []Action {
	available := []Action{}
	for _, action := range actions {
		if (rank < action.Permissions || action.Permissions == NoPerms) ||
			(noJSON && action.JSONoutput == AlwaysJSON) {
			continue
		}
		available = append(available, action)
	}
	return available
}

func getStaffActions(writer http.ResponseWriter, request *http.Request, staff *gcsql.Staff, wantsJSON bool, infoEv *zerolog.Event, errEv *zerolog.Event) (interface{}, error) {
	availableActions := getAvailableActions(staff.Rank, false)
	return availableActions, nil
}
