package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/gorilla/mux"
	"golang.org/x/exp/maps"
)

type TopArtistsResponse struct {
	Rankings   []*history.ArtistRankings `json:"rankings"`
	ArtistData map[string]db.ArtistData  `json:"artist_data"`
	TrackData  map[string]db.TrackData   `json:"track_data"`
}

func (c *StatsController) GetTopArtistsByTimeframe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}
	filter := getFilterParams(r)

	transaction, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer transaction.Commit(ctx)

	// trackIDs := map[string]bool{}
	rankingResults, code, err := history.ArtistStreamRankingsByTimeframe(ctx, transaction, userUUID, filter, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	// for _, result := range rankingResults {
	// 	for _, artistData := range result.Artists {
	// 		for _, trackURI := range artistData.Tracks {
	// 			id, err := service.IDFromURI(trackURI)
	// 			if err != nil {
	// 				fmt.Println(err)
	// 				continue
	// 			}
	// 			trackIDs[id] = true
	// 		}
	// 	}
	// }

	// tracks, err := service.GetTracks(ctx, spClient, maps.Keys(trackIDs))
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	artistIDs := map[string]bool{}
	for _, result := range rankingResults {
		for _, artist := range result.Artists {
			artistIDs[artist.ID] = true
		}
	}

	artistResults, err := service.GetArtists(ctx, spClient, maps.Keys(artistIDs))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := TopArtistsResponse{
		Rankings:   rankingResults,
		ArtistData: artistResults,
		// TrackData:  tracks,
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
	Artist     db.ArtistData           `json:"artist"`
	Streams    []*Stream               `json:"streams"`
	Tracks     map[string]db.TrackData `json:"tracks"`
	TrackRanks []*history.TrackStreams `json:"track_ranks"`
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

	filter := getFilterParams(r)
	log.Printf("%+v", filter)

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer tx.Commit(ctx)

	rows, err := history.AllArtistStreamsByURI(
		ctx, tx, userUUID, artistURI, filter,
	)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	filter.ArtistURIs = []string{artistURI}
	_, _, trackRanks, err := history.CalcTrackStreamsAndRanks(ctx, userUUID, filter, tx, nil, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get track counts: %s", err), http.StatusInternalServerError)
		return
	}

	if trackRanks == nil {
		trackRanks = []*history.TrackStreams{}
	}

	trackIDs := map[string]bool{}
	for _, stream := range trackRanks {
		trackIDs[stream.ID] = true
	}

	streams := []*Stream{}
	for _, row := range rows {
		streams = append(streams, &Stream{
			Timestamp:       row.Timestamp,
			TrackName:       row.TrackName,
			AlbumName:       row.AlbumName,
			MsPlayed:        int(row.MsPlayed),
			SpotifyTrackUri: row.SpotifyTrackUri,
			SpotifyAlbumUri: row.SpotifyAlbumUri,
		})
		id, err := service.IDFromURI(row.SpotifyTrackUri)
		if err != nil {
			continue
		}
		trackIDs[id] = true
	}

	tracks, err := service.GetTracks(ctx, spClient, maps.Keys(trackIDs))
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get streamed tracks: %s", err), http.StatusInternalServerError)
		return
	}

	response := ArtistStatsResponse{
		Streams:    streams,
		Artist:     *artist,
		Tracks:     tracks,
		TrackRanks: trackRanks,
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

	artistURI := mux.Vars(r)["spotify_uri"]
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

	rankings := []TimeframeRanking{}

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer tx.Commit(ctx)

	filter := getFilterParams(r)
	log.Printf("%+v", filter)
	allRankings, responseCode, err := history.ArtistStreamRankingsByTimeframe(ctx, tx, userUUID, filter, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	for _, timeframeRankings := range allRankings {
		for _, artistPlays := range timeframeRankings.Artists {
			if artistPlays.ID == string(artist.ID) {
				ranking := TimeframeRanking{
					Position:             int(artistPlays.Rank),
					Timeframe:            timeframeRankings.Timeframe,
					StartDateUnixSeconds: timeframeRankings.StartDateUnixSeconds,
				}
				rankings = append(rankings, ranking)
			}
		}
	}

	json.NewEncoder(w).Encode(rankings)
}
