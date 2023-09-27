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
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorBadRequest))
		return
	}

	tx, err := db.Service().DB.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("create transaction: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}

	userID, err := db.Service().UserStore.InsertUser(r.Context(), tx, req.Username, req.DisplayName, req.Password)
	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok && pgErr.Code == "23505" {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write(MarshalErrorBody(constants.ErrorUsernameInUse))
			return
		}
		log.Printf("create user: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
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
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
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
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(respBody)
}

func (c *Controller) CurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorNotAuthenticated))
		return
	}

	u, err := db.Service().UserStore.GetByID(r.Context(), userID)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorNotAuthenticated))
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func (c *Controller) GetUserRoom(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorPassword))
		return
	}

	authorized, err := db.Service().UserStore.Authenticate(r.Context(), username, password)
	if err != nil {
		log.Printf("authorize user: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}

	if !authorized {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorPassword))
		return
	}
}
