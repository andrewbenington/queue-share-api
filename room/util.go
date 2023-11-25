package room

import (
	"net/http"

	"github.com/gorilla/mux"
)

// ParametersFromRequest returns the room code from the URL parameter
// and the guest name / room password from the Basic auth header
func ParametersFromRequest(r *http.Request) (code string, guest_id string, password string) {
	vars := mux.Vars(r)
	code = vars["code"]
	guest_id = r.URL.Query().Get("guest_id")
	password = r.URL.Query().Get("password")
	return
}
