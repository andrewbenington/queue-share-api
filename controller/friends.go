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

	suggestions, err := db.New(db.Service().DB).UserGetFriendSuggestions(r.Context(), userUUID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	sent, err := db.New(db.Service().DB).UserGetSentFriendRequests(r.Context(), userUUID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	received, err := db.New(db.Service().DB).UserGetReceivedFriendRequests(r.Context(), userUUID)
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

	err = db.New(db.Service().DB).UserSendFriendRequest(r.Context(), db.UserSendFriendRequestParams{
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

	transaction, err := db.Service().DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	err = db.New(transaction).UserDeleteFriendRequest(r.Context(), db.UserDeleteFriendRequestParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	err = db.New(transaction).UserDeleteFriendRequest(r.Context(), db.UserDeleteFriendRequestParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	err = transaction.Commit()
	if err != nil {
		http.Error(w, "Error committing DB transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) UserAcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
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

	transaction, err := db.Service().DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	ok, err = db.New(transaction).UserGetFriendRequestExists(r.Context(), db.UserGetFriendRequestExistsParams{
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

	err = db.New(transaction).UserDeleteFriendRequest(r.Context(), db.UserDeleteFriendRequestParams{
		UserID:   friendUUID,
		FriendID: userUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	err = db.New(transaction).UserInsertFriend(r.Context(), db.UserInsertFriendParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	err = db.New(transaction).UserInsertFriend(r.Context(), db.UserInsertFriendParams{
		UserID:   friendUUID,
		FriendID: userUUID,
	})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	err = transaction.Commit()
	if err != nil {
		http.Error(w, "Error committing DB transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (c *Controller) UserGetFriends(w http.ResponseWriter, r *http.Request) {
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

	friends, err := db.New(db.Service().DB).UserGetFriends(r.Context(), userUUID)
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

	isFriend, err := db.New(db.Service().DB).UserIsFriends(ctx, db.UserIsFriendsParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	})

	if isFriend && err == nil {
		return friendUUID, nil
	}

	return userUUID, nil
}
