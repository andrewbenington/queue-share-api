package room

import (
	"net/http"

	"github.com/gorilla/mux"
)

func ParametersFromRequest(r *http.Request) (code string, password string) {
	vars := mux.Vars(r)
	code = vars["code"]
	_, password, _ = r.BasicAuth()
	return
}
