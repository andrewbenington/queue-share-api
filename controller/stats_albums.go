package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/google/uuid"
	z_spotify "github.com/zmb3/spotify/v2"
	"golang.org/x/exp/maps"
)

type TopAlbumsResponse struct {
	Rankings  []*history.MonthTopAlbums      `json:"rankings"`
	AlbumData map[string]z_spotify.FullAlbum `json:"album_data"`
}

func (c *Controller) GetTopAlbumsByMonth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, authenticatedAsUser := ctx.Value(auth.UserContextKey).(string)
	if !authenticatedAsUser {
		requests.RespondAuthError(w)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	filter := getFilterParams(r)

	transaction := db.Service().DB

	rankingResults, code, err := history.AlbumStreamRankingsByMonth(ctx, transaction, userID, filter, 30)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	albumIDs := map[string]bool{}
	for _, result := range rankingResults {
		for _, album := range result.Albums {
			log.Printf("%+v\n", album.ID)
			albumIDs[album.ID] = true
		}
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	albumResults, err := spotify.GetAlbums(ctx, spClient, maps.Keys(albumIDs))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := TopAlbumsResponse{
		Rankings:  rankingResults,
		AlbumData: albumResults,
	}
	json.NewEncoder(w).Encode(response)
}
