package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/room"
	"github.com/jackc/pgx/v5/pgconn"
)

func (*Controller) JoinRoomAsMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		requests.RespondWithDBError(w, err)
	}
	defer tx.Commit(ctx)

	reqCtx, err := getRoomRequestContext(ctx, r)
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

	err = room.AddMember(ctx, tx, reqCtx.Room.ID, reqCtx.UserID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	resp := room.RoomResponse{
		Room: *reqCtx.Room,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

type AddMemberRequest struct {
	Username    string `json:"username"`
	IsModerator bool   `json:"is_moderator"`
}

func (*Controller) AddMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		requests.RespondWithDBError(w, err)
	}
	defer tx.Commit(ctx)

	reqCtx, err := getRoomRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Host {
		requests.RespondWithRoomAuthError(w, int(reqCtx.PermissionLevel))
		return
	}

	var body AddMemberRequest
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = room.AddMemberByUsername(ctx, tx, reqCtx.Room.ID, body.Username, body.IsModerator)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			requests.RespondWithError(w, http.StatusConflict, "User already added")
		} else {
			requests.RespondWithDBError(w, err)
		}
		return
	}

	resp, err := getRoomGuestsAndMembers(ctx, reqCtx.Room.ID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type SetModeratorRequest struct {
	UserID      string `json:"user_id"`
	IsModerator bool   `json:"is_moderator"`
}

func (*Controller) SetModerator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		requests.RespondWithDBError(w, err)
	}
	defer tx.Commit(ctx)

	reqCtx, err := getRoomRequestContext(ctx, r)
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

	err = room.SetModerator(ctx, tx, reqCtx.Room.ID, body.UserID, body.IsModerator)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	resp, err := getRoomGuestsAndMembers(ctx, reqCtx.Room.ID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

func (*Controller) DeleteMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		requests.RespondWithDBError(w, err)
	}
	defer tx.Commit(ctx)

	reqCtx, err := getRoomRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Host {
		requests.RespondWithRoomAuthError(w, int(reqCtx.PermissionLevel))
		return
	}

	userID := r.URL.Query().Get("user_id")

	err = room.RemoveMember(ctx, tx, reqCtx.Room.ID, userID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	resp, err := getRoomGuestsAndMembers(ctx, reqCtx.Room.ID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

func (*Controller) AddGuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		requests.RespondWithDBError(w, err)
	}
	defer tx.Commit(ctx)

	reqCtx, err := getRoomRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Guest {
		requests.RespondWithRoomAuthError(w, int(reqCtx.PermissionLevel))
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
		guest, err = room.InsertGuestWithID(ctx, tx, reqCtx.Room.Code, req.Name, reqCtx.GuestID)
	} else {
		guest, err = room.InsertGuest(ctx, tx, reqCtx.Room.Code, req.Name)
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
	reqCtx, err := getRoomRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Guest {
		requests.RespondWithRoomAuthError(w, int(reqCtx.PermissionLevel))
		return
	}

	resp, err := getRoomGuestsAndMembers(ctx, reqCtx.Room.ID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func getRoomGuestsAndMembers(ctx context.Context, roomID string) (*GetRoomGuestsAndMembersResponse, error) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Commit(ctx)

	guests, err := room.GetAllRoomGuests(ctx, tx, roomID)
	if err != nil {
		return nil, err
	}

	members, err := room.GetAllMembers(ctx, tx, roomID)
	if err != nil {
		return nil, err
	}

	return &GetRoomGuestsAndMembersResponse{
		Guests:  guests,
		Members: members,
	}, nil
}
