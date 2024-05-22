package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/exp/maps"
)

type TopArtistsResponse struct {
	Rankings   []*history.MonthTopArtists    `json:"rankings"`
	ArtistData map[string]spotify.FullArtist `json:"artist_data"`
}

func (c *StatsController) GetTopArtistsByTimeframe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}
	filter := getFilterParams(r)

	transaction := db.Service().DB

	rankingResults, code, err := history.ArtistStreamRankingsByTimeframe(ctx, transaction, userUUID, filter, nil, nil)
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

	artistResults, err := service.GetArtists(ctx, spClient, maps.Keys(artistIDs))
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

type AlbumCount struct {
	Name     string  `json:"name"`
	URI      string  `json:"uri"`
	ID       string  `json:"id"`
	ImageURL *string `json:"image_url"`
	Count    int     `json:"count"`
}

type ArtistStatsResponse struct {
	Artist  *spotify.FullArtist     `json:"artist"`
	Streams []*Stream               `json:"streams"`
	Tracks  map[string]db.TrackData `json:"tracks"`
}

func (c *StatsController) GetArtistStatsByURI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
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
	artistID, err := service.IDFromURI(artistURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad artist uri: %s", err), http.StatusBadRequest)
		return
	}

	artist, err := service.GetArtist(ctx, spClient, artistID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get artist: %s", err), http.StatusInternalServerError)
		return
	}

	rows, err := history.AllArtistStreamsByURI(
		ctx,
		db.Service().DB,
		userUUID,
		artistURI,
		getFilterParams(r),
	)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	streams := []*Stream{}
	trackIDs := map[string]bool{}
	for _, row := range rows {
		streams = append(streams, &Stream{
			Timestamp:       row.Timestamp,
			TrackName:       row.TrackName,
			AlbumName:       row.AlbumName,
			MsPlayed:        int(row.MsPlayed),
			SpotifyTrackUri: row.SpotifyTrackUri,
			SpotifyAlbumUri: row.SpotifyAlbumUri.String,
		})
		trackIDs[service.IDFromURIMust(row.SpotifyTrackUri)] = true
	}

	tracks, err := service.GetTracks(ctx, spClient, maps.Keys(trackIDs))
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get streamed tracks: %s", err), http.StatusInternalServerError)
		return
	}

	response := ArtistStatsResponse{
		Streams: streams,
		Artist:  artist,
		Tracks:  tracks,
	}

	json.NewEncoder(w).Encode(response)
}

func (c *StatsController) GetArtistRankingsByURI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
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
	artistID, err := service.IDFromURI(artistURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad artist uri: %s", err), http.StatusBadRequest)
		return
	}

	artist, err := service.GetArtist(ctx, spClient, artistID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get artist: %s", err), http.StatusInternalServerError)
		return
	}

	rankings := []MonthRanking{}

	filter := getFilterParams(r)
	allRankings, responseCode, err := history.ArtistStreamRankingsByTimeframe(ctx, db.Service().DB, userUUID, filter, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	for _, monthRankings := range allRankings {
		for i, artistPlays := range monthRankings.Artists {
			if artistPlays.ID == string(artist.ID) {
				ranking := MonthRanking{
					Year:                 monthRankings.Year,
					Month:                monthRankings.Month,
					Position:             i + 1,
					Timeframe:            monthRankings.Timeframe,
					StartDateUnixSeconds: monthRankings.StartDateUnixSeconds,
				}
				rankings = append(rankings, ranking)
			}
		}
	}

	json.NewEncoder(w).Encode(rankings)
}
