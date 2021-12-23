package manage

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gochan-org/gochan/pkg/config"
	"github.com/gochan-org/gochan/pkg/gclog"
	"github.com/gochan-org/gochan/pkg/gctemplates"
	"github.com/gochan-org/gochan/pkg/gcutil"
	"github.com/gochan-org/gochan/pkg/serverutil"
)

func createNoJSONError(action string) *ErrStaffAction {
	return &ErrStaffAction{
		ErrorField: "nojson",
		Action:     action,
		Message:    "Requested mod page does not have a JSON output option",
	}
}

type ErrStaffAction struct {
	// ErrorField can be used in the frontend for giving more specific info about the error
	ErrorField string `json:"error"`
	Action     string `json:"action"`
	Message    string `json:"message"`
}

func (esa *ErrStaffAction) Error() string {
	return esa.Message
}

func serveError(writer http.ResponseWriter, field string, action string, message string, isJSON bool) {
	if isJSON {
		writer.Header().Add("Content-Type", "application/json")
		writer.Header().Add("Cache-Control", "max-age=5, must-revalidate")
		errJSON, _ := gcutil.MarshalJSON(ErrStaffAction{
			ErrorField: field,
			Action:     action,
			Message:    message,
		}, true)

		serverutil.MinifyWriter(writer, []byte(errJSON), "application/json")
		return
	}
	serverutil.ServeErrorPage(writer, message)
}

func isRequestingJSON(request *http.Request) bool {
	field := request.Form["json"]
	return len(field) == 1 && (field[0] == "1" || field[0] == "true")
}

// CallManageFunction is called when a user accesses /manage to use manage tools
// or log in to a staff account
func CallManageFunction(writer http.ResponseWriter, request *http.Request) {
	var err error
	if err = request.ParseForm(); err != nil {
		serverutil.ServeErrorPage(writer, gclog.Print(gclog.LErrorLog,
			"Error parsing form data: ", err.Error()))
		return
	}
	wantsJSON := isRequestingJSON(request)
	action := request.FormValue("action")
	staffRank := GetStaffRank(request)
	var managePageBuffer bytes.Buffer

	if action == "" {
		if staffRank == NoPerms {
			action = "login"
		} else {
			action = "dashboard"
		}
	}

	handler, ok := actions[action]
	if !ok {
		if wantsJSON {
			serveError(writer, "notfound", action, "action not found", wantsJSON)
		} else {
			serverutil.ServeNotFound(writer, request)
		}
		return
	}
	if action == "actions" {
		handler.Callback = getStaffActions
		wantsJSON = true
	}
	if staffRank == NoPerms && handler.Permissions > NoPerms {
		handler = actions["login"]
	} else if staffRank < handler.Permissions {
		writer.WriteHeader(403)
		staffName, _ := getCurrentStaff(request)

		gclog.Printf(gclog.LStaffLog,
			"Rejected request to manage page %s from %s (insufficient permissions)",
			action, staffName)

		serveError(writer, "permission", action, "You do not have permission to access this page", handler.isJSON || wantsJSON)
		return
	}

	output, err := handler.Callback(writer, request, isRequestingJSON(request))
	if err != nil {
		staffName, _ := getCurrentStaff(request)
		// writer.WriteHeader(500)
		gclog.Printf(gclog.LStaffLog|gclog.LErrorLog,
			"Error accessing manage page %s by %s: %s", action, staffName, err.Error())
		serveError(writer, "actionerror", action, err.Error(), wantsJSON || handler.isJSON)
		return
	}
	if handler.isJSON || wantsJSON {
		writer.Header().Add("Content-Type", "application/json")
		writer.Header().Add("Cache-Control", "max-age=5, must-revalidate")
		outputJSON, err := gcutil.MarshalJSON(output, true)
		if err != nil {
			serveError(writer, "error", action, err.Error(), true)
			return
		}
		serverutil.MinifyWriter(writer, []byte(outputJSON), "application/json")
		return
	}
	managePageBuffer.WriteString("<!DOCTYPE html><html><head>")
	criticalCfg := config.GetSystemCriticalConfig()

	if err = serverutil.MinifyTemplate(gctemplates.ManageHeader,
		map[string]interface{}{
			"webroot": criticalCfg.WebRoot,
		},
		&managePageBuffer, "text/html"); err != nil {
		serverutil.ServeErrorPage(writer, gclog.Print(gclog.LErrorLog|gclog.LStaffLog,
			"Error executing manage page header template: ", err.Error()))
		return
	}
	managePageBuffer.WriteString(fmt.Sprint(output, "</body></html>"))
	writer.Write(managePageBuffer.Bytes())
}
