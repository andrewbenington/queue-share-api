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
	"github.com/andrewbenington/queue-share-api/constants"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/room"
)

type PermissionLevel int

type RequestContext struct {
	Room            *room.Room
	UserID          string
	GuestID         string
	PermissionLevel PermissionLevel
}

const (
	NotAuthorized PermissionLevel = iota
	Forbidden
	Guest
	Member
	Moderator
	Host
)

func (c *Controller) CreateRoom(w http.ResponseWriter, r *http.Request) {
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

	newRoom, err := db.Service().RoomStore.Insert(r.Context(), req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error inserting room: %s", err)
		requests.RespondInternalError(w)
		return
	}

	user, err := db.Service().UserStore.GetByID(r.Context(), userID)
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
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func (c *Controller) GetRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Guest {
		requests.RespondAuthError(w)
		return
	}

	resp := room.RoomResponse{
		Room: *reqCtx.Room,
	}

	if reqCtx.GuestID != "" {
		guestName, err := db.Service().RoomStore.GetGuestName(ctx, reqCtx.Room.ID, reqCtx.GuestID)
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

	reqCtx, err := getRequestContext(ctx, r)
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

func (*Controller) JoinRoomAsMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel >= Member {
		w.WriteHeader(http.StatusNotModified)
		_, _ = w.Write([]byte{})
		return
	}

	if reqCtx.UserID == "" || reqCtx.PermissionLevel < Guest {
		requests.RespondAuthError(w)
		return
	}

	err = db.Service().RoomStore.AddMember(ctx, reqCtx.Room.Code, reqCtx.UserID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte{})
}

func (*Controller) AddGuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Guest {
		requests.RespondAuthError(w)
		return
	}

	var req room.InsertGuestRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		requests.RespondBadRequest(w)
		return
	}

	var guest *room.Guest
	if reqCtx.GuestID != "" {
		guest, err = db.Service().RoomStore.InsertGuestWithID(ctx, reqCtx.Room.Code, req.Name, reqCtx.GuestID)
	} else {
		guest, err = db.Service().RoomStore.InsertGuest(ctx, reqCtx.Room.Code, req.Name)
	}

	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	body, err := json.Marshal(guest)
	if err != nil {
		log.Printf("error marshalling room guest: %s\n", err)
		requests.RespondInternalError(w)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

type GetRoomGuestsAndMembersResponse struct {
	Guests  []room.Guest  `json:"guests"`
	Members []room.Member `json:"members"`
}

func (*Controller) GetRoomGuestsAndMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqCtx, err := getRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	log.Printf("%+v\n", reqCtx)
	if reqCtx.PermissionLevel < Guest {
		requests.RespondAuthError(w)
		return
	}

	guests, err := db.Service().RoomStore.GetAllRoomGuests(ctx, reqCtx.Room.Code)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	members, err := db.Service().RoomStore.GetAllMembers(ctx, reqCtx.Room.ID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	body, err := json.Marshal(GetRoomGuestsAndMembersResponse{
		Guests:  guests,
		Members: members,
	})
	if err != nil {
		log.Printf("error marshalling room guest: %s\n", err)
		requests.RespondWithError(w, http.StatusInternalServerError, constants.ErrorInternal)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (*Controller) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code, _, _ := room.ParametersFromRequest(r)
	if code == "" {
		requests.RespondBadRequest(w)
		return
	}

	hostID, err := db.Service().RoomStore.GetHostID(ctx, code)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	userID, ok := r.Context().Value(auth.UserContextKey).(string)

	if !ok || hostID != userID {
		requests.RespondAuthError(w)
		return
	}

	err = db.Service().RoomStore.DeleteByCode(ctx, code)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte{})
}

type SetModeratorRequest struct {
	UserID      string `json:"user_id"`
	IsModerator bool   `json:"is_moderator"`
}

func (*Controller) SetModerator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqCtx, err := getRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Host {
		requests.RespondAuthError(w)
		return
	}

	var body SetModeratorRequest
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.Service().RoomStore.SetModerator(ctx, reqCtx.Room.ID, body.UserID, body.IsModerator)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getRequestContext(ctx context.Context, r *http.Request) (as RequestContext, err error) {
	code, guestID, password := room.ParametersFromRequest(r)
	rm, err := db.Service().RoomStore.GetByCode(ctx, code)
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

		isModerator, err := db.Service().RoomStore.UserIsMember(ctx, rm.ID, userID)

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

	passwordValid, err := db.Service().RoomStore.ValidatePassword(ctx, code, password)
	if err != nil && err != sql.ErrNoRows {
		return as, err
	}
	if !passwordValid {
		as.PermissionLevel = NotAuthorized
		return as, nil
	}
	as.PermissionLevel = Guest
	return as, nil
}
