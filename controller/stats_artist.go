package controller

import (
	"encoding/json"
	"fmt"
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

type TopArtistsResponse struct {
	Rankings   []*history.MonthTopArtists      `json:"rankings"`
	ArtistData map[string]z_spotify.FullArtist `json:"artist_data"`
}

func (c *Controller) GetTopArtistsByMonth(w http.ResponseWriter, r *http.Request) {
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

	rankingResults, code, err := history.ArtistStreamRankingsByMonth(ctx, transaction, userID, filter, 30)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	artistIDs := map[string]bool{}
	for _, result := range rankingResults {
		for _, artist := range result.Artists {
			artistIDs[artist.ID] = true
		}
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	artistResults, err := spotify.GetArtists(ctx, spClient, maps.Keys(artistIDs))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := TopArtistsResponse{
		Rankings:   rankingResults,
		ArtistData: artistResults,
	}

	json.NewEncoder(w).Encode(response)
}

type ArtistStatsResponse struct {
	Artist   *z_spotify.FullArtist `json:"artist"`
	Streams  []*Stream             `json:"streams"`
	Rankings []MonthRanking        `json:"rankings"`
}

func (c *Controller) GetArtistStatsByURI(w http.ResponseWriter, r *http.Request) {
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

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get client: %s", err), code)
		return
	}

	artistURI := r.URL.Query().Get("spotify_uri")
	artistID, err := spotify.IDFromURI(artistURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad artist uri: %s", err), http.StatusBadRequest)
		return
	}

	artist, err := spotify.GetArtist(ctx, spClient, artistID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get artist: %s", err), http.StatusInternalServerError)
		return
	}

	rows, err := history.AllArtistStreamsByURI(
		ctx,
		db.Service().DB,
		userID,
		artistURI,
		getFilterParams(r),
	)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}
	streams := []*Stream{}
	for _, row := range rows {
		streams = append(streams, &Stream{
			Timestamp:        row.Timestamp,
			TrackName:        row.TrackName,
			ArtistName:       row.ArtistName,
			AlbumName:        row.AlbumName,
			MsPlayed:         int(row.MsPlayed),
			SpotifyTrackUri:  row.SpotifyTrackUri,
			SpotifyArtistUri: row.SpotifyArtistUri.String,
			SpotifyAlbumUri:  row.SpotifyAlbumUri.String,
		})
	}
	response := ArtistStatsResponse{
		Streams:  streams,
		Artist:   artist,
		Rankings: []MonthRanking{},
	}

	filter := getFilterParams(r)
	allRankings, responseCode, err := history.ArtistStreamRankingsByMonth(ctx, db.Service().DB, userID, filter, 30)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	for _, monthRankings := range allRankings {
		for i, artistPlays := range monthRankings.Artists {
			if artistPlays.ID == string(artist.ID) {
				ranking := MonthRanking{
					Year:     monthRankings.Year,
					Month:    monthRankings.Month,
					Position: i + 1,
				}
				response.Rankings = append(response.Rankings, ranking)
			}
		}
	}

	json.NewEncoder(w).Encode(response)
}
