package requests

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewbenington/queue-share-api/constants"
)

func RespondWithError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_, _ = w.Write(marshalErrorBody(message))
}

func RespondWithDBError(w http.ResponseWriter, err error) {
	if err == sql.ErrNoRows {
		RespondNotFound(w)
		return
	}
	log.Printf("db error: %s\n", err)
	RespondInternalError(w)
}

func RespondWithRoomAuthError(w http.ResponseWriter, permissionLevel int) {
	if permissionLevel == 0 {
		RespondWithError(w, http.StatusUnauthorized, constants.ErrorPassword)
		return
	}
	RespondWithError(w, http.StatusForbidden, constants.ErrorForbidden)
}

func RespondNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(marshalErrorBody(constants.ErrorNotFound))
}

func RespondBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write(marshalErrorBody(constants.ErrorBadRequest))
}

func RespondInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(marshalErrorBody(constants.ErrorInternal))
}

func RespondAuthError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write(marshalErrorBody(constants.ErrorNotAuthenticated))
}

func marshalErrorBody(e string) []byte {
	body, err := json.MarshalIndent(ErrorResponse{Error: e}, "", " ")
	if err != nil {
		body, _ = json.MarshalIndent(ErrorResponse{Error: err.Error()}, "", " ")
	}
	return body
}

type ErrorResponse struct {
	Error string `json:"error"`
}
