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

type TopAlbumsResponse struct {
	Rankings  []*history.MonthTopAlbums    `json:"rankings"`
	AlbumData map[string]spotify.FullAlbum `json:"album_data"`
}

func (c *StatsController) GetTopAlbumsByMonth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	filter := getFilterParams(r)

	transaction := db.Service().DB

	rankingResults, code, err := history.AlbumStreamRankingsByMonth(ctx, transaction, userUUID, filter, 30)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	albumIDs := map[string]bool{}
	for _, result := range rankingResults {
		for _, album := range result.Albums {
			albumIDs[album.ID] = true
		}
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	albumResults, err := service.GetAlbums(ctx, spClient, maps.Keys(albumIDs))
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

type AlbumStatsResponse struct {
	Album   *spotify.FullAlbum      `json:"album"`
	Streams []*Stream               `json:"streams"`
	Tracks  map[string]db.TrackData `json:"tracks"`
}

func (c *StatsController) GetAlbumStatsByURI(w http.ResponseWriter, r *http.Request) {
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

	albumURI := r.URL.Query().Get("spotify_uri")
	albumID, err := service.IDFromURI(albumURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad album uri: %s", err), http.StatusBadRequest)
		return
	}

	album, err := service.GetAlbum(ctx, spClient, albumID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get album: %s", err), http.StatusInternalServerError)
		return
	}

	rows, err := history.AllAlbumStreamsByURI(
		ctx,
		db.Service().DB,
		userUUID,
		albumURI,
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
		})
		trackIDs[service.IDFromURIMust(row.SpotifyTrackUri)] = true
	}

	tracks, err := service.GetTracks(ctx, spClient, maps.Keys(trackIDs))
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get streamed tracks: %s", err), http.StatusInternalServerError)
		return
	}

	response := AlbumStatsResponse{
		Streams: streams,
		Album:   album,
		Tracks:  tracks,
	}

	json.NewEncoder(w).Encode(response)
}

func (c *StatsController) GetAlbumRankingsByURI(w http.ResponseWriter, r *http.Request) {
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

	albumURI := r.URL.Query().Get("spotify_uri")
	albumID, err := service.IDFromURI(albumURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad album uri: %s", err), http.StatusBadRequest)
		return
	}

	album, err := service.GetAlbum(ctx, spClient, albumID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get album: %s", err), http.StatusInternalServerError)
		return
	}

	rankings := []MonthRanking{}

	filter := getFilterParams(r)
	allRankings, responseCode, err := history.AlbumStreamRankingsByMonth(ctx, db.Service().DB, userUUID, filter, 30)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	for _, monthRankings := range allRankings {
		for i, albumPlays := range monthRankings.Albums {
			if albumPlays.ID == string(album.ID) {
				ranking := MonthRanking{
					Year:     monthRankings.Year,
					Month:    monthRankings.Month,
					Position: i + 1,
				}
				rankings = append(rankings, ranking)
			}
		}
	}

	json.NewEncoder(w).Encode(rankings)
}
