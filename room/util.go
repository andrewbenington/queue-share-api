package room

import (
	"net/http"

	"github.com/gorilla/mux"
)

func ParametersFromRequest(r *http.Request) (code string, user string, password string) {
	vars := mux.Vars(r)
	code = vars["code"]
	user, password, _ = r.BasicAuth()
	return
}
