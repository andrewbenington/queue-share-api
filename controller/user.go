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

	userID, err := db.Service().UserStore.InsertUser(r.Context(), tx, req.Username, req.DisplayName, req.Password)
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

	u, err := db.Service().UserStore.GetByID(r.Context(), userID)

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

type GetUserRoomResponse struct {
	Room *user.GetUserRoomResponse `json:"room"`
}

func (c *Controller) GetUserRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	room, err := db.Service().UserStore.GetUserRoom(r.Context(), userID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error getting user room: %s", err)
		requests.RespondInternalError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetUserRoomResponse{Room: room})
}

func (c *Controller) UnlinkSpotify(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	err := db.Service().UserStore.UnlinkSpotify(r.Context(), userID)
	if err != nil {
		requests.RespondWithDBError(w, err)
	}
	w.WriteHeader(http.StatusNoContent)
}
