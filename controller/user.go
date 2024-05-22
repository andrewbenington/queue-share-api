package controller

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/constants"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/room"
	"github.com/andrewbenington/queue-share-api/user"
	"github.com/jackc/pgx/v5/pgconn"
)

type CreateUserBody struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type CreateUserResponse struct {
	User      user.User `json:"user"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserBody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		requests.RespondBadRequest(w)
		return
	}

	tx, err := db.Service().DB.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("create transaction: %s", err)
		requests.RespondInternalError(w)
		return
	}

	userID, err := user.InsertUser(r.Context(), tx, req.Username, req.DisplayName, req.Password)
	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok && pgErr.Code == "23505" {
			requests.RespondWithError(w, http.StatusConflict, constants.ErrorUsernameInUse)
			return
		}
		log.Printf("create user: %s", err)
		requests.RespondInternalError(w)
		return
	}

	user := user.User{
		ID:          userID,
		Username:    req.Username,
		DisplayName: req.DisplayName,
	}

	token, expiry, err := user.GetJWT()
	if err != nil {
		_ = tx.Rollback()
		log.Printf("generate jwt: %s", err)
		requests.RespondInternalError(w)
		return
	}

	respBody := CreateUserResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiry,
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("commit tx: %s", err)
		requests.RespondInternalError(w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(respBody)
}

func (c *Controller) CurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	u, err := user.GetByID(r.Context(), db.Service().DB, userID)

	if err == sql.ErrNoRows {
		requests.RespondAuthError(w)
		return
	}
	if err != nil {
		log.Printf("Error getting current user: %s", err)
		requests.RespondInternalError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

type GetRoomsResponse struct {
	Rooms []room.Room `json:"rooms"`
}

func (c *Controller) GetUserHostedRooms(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	rooms, err := room.GetUserHostedRooms(r.Context(), db.Service().DB, userID, true)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error getting user room: %s", err)
		requests.RespondInternalError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetRoomsResponse{Rooms: rooms})
}

func (c *Controller) GetUserJoinedRooms(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	rooms, err := room.GetUserJoinedRooms(r.Context(), db.Service().DB, userID, true)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error getting user room: %s", err)
		requests.RespondInternalError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetRoomsResponse{Rooms: rooms})
}

func (c *Controller) UnlinkSpotify(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	err := user.UnlinkSpotify(r.Context(), db.Service().DB, userID)
	if err != nil {
		requests.RespondWithDBError(w, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) UserHasSpotifyHistory(w http.ResponseWriter, r *http.Request) {

	userUUID, err := userUUIDFromRequest(r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	result, err := db.New((db.Service().DB)).UserHasSpotifyHistory(r.Context(), userUUID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	response := map[string]bool{"user_has_history": result}
	json.NewEncoder(w).Encode(response)
}
