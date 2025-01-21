package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/andrewbenington/queue-share-api/admin"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/engine"
	"github.com/andrewbenington/queue-share-api/user"
	"github.com/andrewbenington/queue-share-api/util"
	"github.com/samber/lo"
)

func (c *Controller) GetTableData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer tx.Rollback(ctx)

	rows, err := admin.GetTableSizes(ctx, tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rows)
}

func (c *Controller) GetUncachedTracks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer tx.Rollback(ctx)

	rows, err := db.New(tx).UncachedTracks(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rows)
}

func (c *Controller) GetMissingISRCNumbers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer tx.Rollback(ctx)

	rows, err := db.New(tx).MissingISRCNumbers(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rows)
}

func (c *Controller) GetMissingArtistURIs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer tx.Rollback(ctx)

	rows, err := db.New(tx).MissingArtistURIs(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rows)
}

func (c *Controller) GetMissingArtistURIsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer tx.Rollback(ctx)

	missingByUser := engine.GetMissingURIsByUser()

	usersByID := map[string]*user.User{}
	for _, userID := range lo.Keys(missingByUser) {
		usersByID[userID], err = user.GetByID(ctx, tx, userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not get user with id '%s'", userID), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"missing_by_user": missingByUser,
		"users":           usersByID,
	})
}

type FullLogEntry struct {
	Timestamp time.Time  `json:"timestamp"`
	User      *user.User `json:"user"`
	Endpoint  string     `json:"endpoint"`
	Friend    *user.User `json:"friend"`
}

func (c *Controller) GetLogsByDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer tx.Rollback(ctx)

	dateString := r.URL.Query().Get("date")
	timestamp, err := time.Parse("01-02-06", dateString)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad date format: %s", err), http.StatusBadRequest)
		return
	}

	logEntries, err := util.GetLogsForDate(timestamp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fullLogEntries := []FullLogEntry{}
	usersByID := map[string]user.User{}

	for _, entry := range logEntries {
		userData, err := userFromDBOrMap(ctx, entry.UserID, usersByID, tx)
		if err != nil {
			log.Printf("User %s not found", entry.UserID)
			continue
		}
		var friend *user.User
		if entry.FriendID != "" {
			friend, err = userFromDBOrMap(ctx, entry.FriendID, usersByID, tx)
			if err != nil {
				log.Printf("Friend %s not found", entry.UserID)
				continue
			}
		}

		fullLogEntries = append(fullLogEntries, FullLogEntry{
			Timestamp: entry.Timestamp,
			User:      userData,
			Endpoint:  entry.Endpoint,
			Friend:    friend,
		})
	}

	json.NewEncoder(w).Encode(fullLogEntries)
}

func (c *Controller) GetGeneralInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer tx.Rollback(ctx)
	resp := map[string]interface{}{}
	resp["log_queue"] = util.GetLogChannelSize()
	json.NewEncoder(w).Encode(resp)
}

func userFromDBOrMap(ctx context.Context, userID string, userMap map[string]user.User, tx db.DBTX) (*user.User, error) {
	if userData, ok := userMap[userID]; ok {
		return &userData, nil
	}

	return user.GetByID(ctx, tx, userID)
}
