package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/google/uuid"
)

type FriendRequestData struct {
	Suggestions      []*db.UserGetFriendSuggestionsRow      `json:"suggestions"`
	SentRequests     []*db.UserGetSentFriendRequestsRow     `json:"sent_requests"`
	ReceivedRequests []*db.UserGetReceivedFriendRequestsRow `json:"received_requests"`
}

func (c *Controller) UserFriendRequestData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	userUUID, err := userUUIDFromRequest(r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		requests.RespondWithDBError(w, err)
	}
	defer tx.Commit(ctx)

	suggestions, err := db.New(tx).UserGetFriendSuggestions(ctx, userUUID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	sent, err := db.New(tx).UserGetSentFriendRequests(ctx, userUUID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	received, err := db.New(tx).UserGetReceivedFriendRequests(ctx, userUUID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	json.NewEncoder(w).Encode(FriendRequestData{
		Suggestions:      suggestions,
		SentRequests:     sent,
		ReceivedRequests: received,
	})
}

func (c *Controller) UserSendFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	friendID := r.URL.Query().Get("friend_id")
	friendUUID, err := uuid.Parse(friendID)
	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	userUUID, err := userUUIDFromRequest(r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		requests.RespondWithDBError(w, err)
	}
	defer tx.Commit(ctx)

	err = db.New(tx).UserSendFriendRequest(r.Context(), db.UserSendFriendRequestParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (c *Controller) UserDeleteFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	friendID := r.URL.Query().Get("friend_id")
	friendUUID, err := uuid.Parse(friendID)
	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	userUUID, err := userUUIDFromRequest(r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	err = db.New(tx).UserDeleteFriendRequest(r.Context(), db.UserDeleteFriendRequestParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	err = db.New(tx).UserDeleteFriendRequest(r.Context(), db.UserDeleteFriendRequestParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, "Error committing DB transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) UserAcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	friendID := r.URL.Query().Get("friend_id")
	friendUUID, err := uuid.Parse(friendID)
	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	userUUID, err := userUUIDFromRequest(r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	ok, err = db.New(tx).UserGetFriendRequestExists(r.Context(), db.UserGetFriendRequestExistsParams{
		FriendID: userUUID,
		UserID:   friendUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}
	if !ok {
		http.Error(w, "Friend request does not exist", http.StatusBadRequest)
		return
	}

	err = db.New(tx).UserDeleteFriendRequest(r.Context(), db.UserDeleteFriendRequestParams{
		UserID:   friendUUID,
		FriendID: userUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	err = db.New(tx).UserInsertFriend(r.Context(), db.UserInsertFriendParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	err = db.New(tx).UserInsertFriend(r.Context(), db.UserInsertFriendParams{
		UserID:   friendUUID,
		FriendID: userUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, "Error committing DB transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (c *Controller) UserGetFriends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	userUUID, err := userUUIDFromRequest(r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer tx.Commit(ctx)

	friends, err := db.New(tx).UserGetFriends(r.Context(), userUUID)
	if err == sql.ErrNoRows {
		friends = []*db.User{}
	} else if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	json.NewEncoder(w).Encode(friends)
}

func userUUIDFromRequest(r *http.Request) (uuid.UUID, error) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("user credentials not present")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse user UUID: %w", err)
	}

	return userUUID, nil
}

func userOrFriendUUIDFromRequest(ctx context.Context, r *http.Request) (uuid.UUID, error) {
	userUUID, err := userUUIDFromRequest(r)
	if err != nil {
		return uuid.UUID{}, err
	}

	friendID := r.URL.Query().Get("friend_id")
	friendUUID, err := uuid.Parse(friendID)
	if err != nil {
		return userUUID, nil
	}

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer tx.Commit(ctx)

	isFriend, err := db.New(tx).UserIsFriends(ctx, db.UserIsFriendsParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	})

	if isFriend && err == nil {
		return friendUUID, nil
	}

	return userUUID, nil
}
