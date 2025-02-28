package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/room"
	"github.com/andrewbenington/queue-share-api/user"
)

type PermissionLevel int

type RequestContext struct {
	Room            *room.Room
	UserID          string
	GuestID         string
	PermissionLevel PermissionLevel
}

const (
	IncorrectPassword PermissionLevel = iota
	NotAuthorized
	Guest
	Member
	Moderator
	Host
)

func (c *Controller) CreateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	var req room.InsertRoomParams
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		requests.RespondBadRequest(w)
		return
	}

	req.HostID = userID

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	newRoom, err := room.Insert(ctx, tx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error inserting room: %s", err)
		requests.RespondInternalError(w)
		return
	}

	user, err := user.GetByID(ctx, tx, userID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get room host after create: %s", err)
	} else {
		newRoom.Host = *user
	}

	body, err := json.Marshal(room.RoomResponse{Room: newRoom})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling room: %s", err)
		requests.RespondInternalError(w)
		return
	}
	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, "Error committing DB transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func (c *Controller) GetRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	reqCtx, err := getRoomRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Guest {
		requests.RespondWithError(w, http.StatusUnauthorized, "Password incorrect")
		return
	}

	resp := room.RoomResponse{
		Room: *reqCtx.Room,
	}

	if reqCtx.GuestID != "" {
		guestName, err := room.GetGuestName(ctx, tx, reqCtx.Room.ID, reqCtx.GuestID)
		if err != nil {
			log.Printf("error finding guest name: %s\n", err)
		} else {
			resp.Guest = &room.Guest{
				ID:   reqCtx.GuestID,
				Name: guestName,
			}
		}
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Printf("error marshalling room response: %s\n", err)
		requests.RespondInternalError(w)
		return
	}
	w.Write(body)
}

type GetRoomAuthLevelResponse struct {
	IsMember    bool `json:"is_member"`
	IsModerator bool `json:"is_moderator"`
	IsHost      bool `json:"is_host"`
}

func (c *Controller) GetRoomPermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRoomRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.UserID == "" {
		requests.RespondAuthError(w)
		return
	}

	var response GetRoomAuthLevelResponse
	switch reqCtx.PermissionLevel {
	case Host:
		response.IsHost = true
		fallthrough
	case Moderator:
		response.IsModerator = true
		fallthrough
	case Member:
		response.IsMember = true
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (*Controller) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	reqCtx, err := getRoomRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Host {
		requests.RespondWithRoomAuthError(w, int(reqCtx.PermissionLevel))
		return
	}

	err = room.DeleteByCode(ctx, tx, reqCtx.Room.Code)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte{})
}

type UpdatePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func (*Controller) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	reqCtx, err := getRoomRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Host {
		requests.RespondAuthError(w)
		return
	}

	var body UpdatePasswordRequest
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = room.UpdatePassword(ctx, tx, reqCtx.Room.ID, body.NewPassword)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getRoomRequestContext(ctx context.Context, r *http.Request) (as RequestContext, err error) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return as, nil
	}

	code, guestID, password := room.ParametersFromRequest(r)
	rm, err := room.GetByCode(ctx, tx, code)
	if err != nil {
		return
	}

	as.Room = &rm

	userID, authenticatedAsUser := r.Context().Value(auth.UserContextKey).(string)
	if authenticatedAsUser {
		as.UserID = userID
		if userID == rm.Host.ID {
			as.PermissionLevel = Host
			return
		}

		isModerator, err := room.UserIsMember(ctx, tx, rm.ID, userID)

		if err == nil {
			if isModerator {
				as.PermissionLevel = Moderator
			} else {
				as.PermissionLevel = Member
			}
			return as, nil
		}

		if err != sql.ErrNoRows {
			return as, err
		}
	}

	// Request has been made by a guest; authenticate with room password
	as.GuestID = guestID

	passwordValid, err := room.ValidatePassword(ctx, tx, code, password)
	if err != nil && err != sql.ErrNoRows {
		return as, err
	}
	if !passwordValid {
		as.PermissionLevel = IncorrectPassword
		return as, nil
	}
	as.PermissionLevel = Guest

	return as, nil
}
