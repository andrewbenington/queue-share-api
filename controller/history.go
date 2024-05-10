package controller

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/engine"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/google/uuid"
	z_spotify "github.com/zmb3/spotify/v2"
	"golang.org/x/exp/maps"
)

const (
	DEFAULT_MIN_MS_FILTER = 30000
	DEFAULT_LIMIT         = 50
)

type HistoryEntry struct {
	Timestamp       time.Time        `json:"timestamp"`
	TrackName       string           `json:"track_name"`
	AlbumName       string           `json:"album_name"`
	MsPlayed        int32            `json:"ms_played"`
	SpotifyTrackUri string           `json:"spotify_track_uri"`
	SpotifyAlbumUri string           `json:"spotify_album_uri"`
	ImageURL        *string          `json:"image_url"`
	Artists         []db.TrackArtist `json:"artists"`
}

type HistoryEntryArtist struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type HistoryResponse struct {
	History     []HistoryEntry `json:"history"`
	LastFetched *time.Time     `json:"last_fetched"`
}

func (c *Controller) GetAllHistory(w http.ResponseWriter, r *http.Request) {
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

	minMSPlayedParam := r.URL.Query().Get("minimum_milliseconds")
	minMSPlayed, err := strconv.Atoi(minMSPlayedParam)
	if err != nil {
		minMSPlayed = DEFAULT_MIN_MS_FILTER
	}

	limitParam := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit > DEFAULT_LIMIT {
		limit = DEFAULT_LIMIT
	}

	includeSkippedParam := r.URL.Query().Get("include_skipped")
	includeSkipped := strings.EqualFold(includeSkippedParam, "true")

	rows, err := db.New(db.Service().DB).HistoryGetAll(ctx, db.HistoryGetAllParams{
		UserID:       userUUID,
		MinMsPlayed:  int32(minMSPlayed),
		IncludeSkips: includeSkipped,
		MaxCount:     int32(limit)})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	trackURIs := map[string]bool{}
	for _, row := range rows {
		trackURIs[row.SpotifyTrackUri] = true
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	trackIDs := []string{}
	for _, uri := range maps.Keys(trackURIs) {
		trackIDs = append(trackIDs, spotify.IDFromURIMust(uri))
	}

	trackByID, err := spotify.GetTracks(ctx, spClient, trackIDs)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	entries := []HistoryEntry{}
	for _, row := range rows {
		entry := HistoryEntry{
			Timestamp:       row.Timestamp,
			TrackName:       row.TrackName,
			SpotifyTrackUri: row.SpotifyTrackUri,
			AlbumName:       row.AlbumName,
			SpotifyAlbumUri: row.SpotifyAlbumUri.String,
			MsPlayed:        row.MsPlayed,
			Artists: []db.TrackArtist{
				{
					Name: row.ArtistName,
					URI:  row.SpotifyArtistUri.String,
					ID:   spotify.IDFromURIMust(row.SpotifyArtistUri.String),
				},
			},
		}

		if track, ok := trackByID[spotify.IDFromURIMust(row.SpotifyTrackUri)]; ok {
			entry.ImageURL = track.ImageUrl
			entry.Artists = append(entry.Artists, track.OtherArtists...)
		}

		entries = append(entries, entry)
	}

	json.NewEncoder(w).Encode(HistoryResponse{
		History:     entries,
		LastFetched: engine.LastFetch,
	})
}

func (c *Controller) UploadHistory(w http.ResponseWriter, r *http.Request) {
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

	zipData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	tempFile, err := os.CreateTemp("", "history-zip-*.zip")
	if err != nil {
		http.Error(w, "Error creating temporary file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	_, err = tempFile.Write(zipData)
	if err != nil {
		http.Error(w, "Error writing temp file", http.StatusInternalServerError)
		return
	}

	zipReader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		http.Error(w, "Error opening zip file", http.StatusInternalServerError)
		return
	}
	defer zipReader.Close()

	transaction, err := db.Service().DB.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	for _, file := range zipReader.File {
		if strings.EqualFold(path.Base(file.Name), "Userdata.json") {
			http.Error(w, "This is the wrong file. Please upload your \"Extended Streaming History\", NOT your \"Account Data\".", http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(path.Base(file.Name), "Streaming_History_Audio") {
			continue
		}

		// Open the file from the zip archive
		zippedFile, err := file.Open()
		if err != nil {
			http.Error(w, "Error opening file from zip", http.StatusInternalServerError)
			return
		}
		defer zippedFile.Close()

		entryData := []history.StreamingEntry{}
		err = json.NewDecoder(zippedFile).Decode(&entryData)
		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
			return
		}

		err = history.InsertEntriesFromHistory(ctx, transaction, userUUID, entryData)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error uploading history", http.StatusInternalServerError)
			return
		}

		fmt.Printf("Uploaded file %s\n", file.Name)
	}
	err = transaction.Commit()
	if err != nil {
		http.Error(w, "Error committing DB transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (c *Controller) GetTopTracksByYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, authenticatedAsUser := ctx.Value(auth.UserContextKey).(string)
	if !authenticatedAsUser {
		requests.RespondAuthError(w)
		return
	}

	filter := getFilterParams(r)

	results, responseCode, err := history.TrackStreamCountByYear(ctx, db.Service().DB, userID, filter)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	json.NewEncoder(w).Encode(results)
}

func (c *Controller) GetTopAlbumsByYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, authenticatedAsUser := ctx.Value(auth.UserContextKey).(string)
	if !authenticatedAsUser {
		requests.RespondAuthError(w)
		return
	}

	filter := getFilterParams(r)

	results, responseCode, err := history.AlbumStreamCountByYear(ctx, db.Service().DB, userID, filter)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	json.NewEncoder(w).Encode(results)
}

func (c *Controller) GetTopArtistsByYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, authenticatedAsUser := ctx.Value(auth.UserContextKey).(string)
	if !authenticatedAsUser {
		requests.RespondAuthError(w)
		return
	}

	filter := getFilterParams(r)

	results, responseCode, err := history.ArtistStreamCountByYear(ctx, db.Service().DB, userID, filter)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	json.NewEncoder(w).Encode(results)
}

func (c *Controller) GetAllStreamsByURI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, authenticatedAsUser := ctx.Value(auth.UserContextKey).(string)
	if !authenticatedAsUser {
		requests.RespondAuthError(w)
		return
	}

	rows, err := history.AllTrackStreamsByURI(
		ctx,
		db.Service().DB,
		userID,
		r.URL.Query().Get("spotify_uri"),
		getFilterParams(r),
	)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	json.NewEncoder(w).Encode(rows)
}

type MonthRanking struct {
	Year     int `json:"year"`
	Month    int `json:"month"`
	Position int `json:"position"`
}

type Stream struct {
	Timestamp        time.Time `json:"timestamp"`
	TrackName        string    `json:"track_name"`
	ArtistName       string    `json:"artist_name"`
	AlbumName        string    `json:"album_name"`
	MsPlayed         int       `json:"ms_played"`
	SpotifyTrackUri  string    `json:"spotify_track_uri"`
	SpotifyArtistUri string    `json:"spotify_artist_uri"`
	SpotifyAlbumUri  string    `json:"spotify_album_uri"`
}
type TrackStatsResponse struct {
	Track    *db.TrackData  `json:"track"`
	Streams  []*Stream      `json:"streams"`
	Rankings []MonthRanking `json:"rankings"`
}

func (c *Controller) GetTrackStatsByURI(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), code)
		return
	}

	trackURI := r.URL.Query().Get("spotify_uri")
	trackID, err := spotify.IDFromURI(trackURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	track, err := spotify.GetTrack(ctx, spClient, trackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := history.AllTrackStreamsByURI(
		ctx,
		db.Service().DB,
		userID,
		trackURI,
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

	response := TrackStatsResponse{
		Streams:  streams,
		Track:    track,
		Rankings: []MonthRanking{},
	}

	filter := getFilterParams(r)
	filter.MaxTracks = 30
	allRankings, responseCode, err := history.TrackStreamRankingsByMonth(ctx, db.Service().DB, userID, filter)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	for _, monthRankings := range allRankings {
		for i, trackPlays := range monthRankings.Tracks {
			if trackPlays.ID == string(track.ID) {
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

type AlbumStatsResponse struct {
	Album    *z_spotify.FullAlbum `json:"album"`
	Streams  []*Stream            `json:"streams"`
	Rankings []MonthRanking       `json:"rankings"`
}

func (c *Controller) GetAlbumStatsByURI(w http.ResponseWriter, r *http.Request) {
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

	albumURI := r.URL.Query().Get("spotify_uri")
	albumID, err := spotify.IDFromURI(albumURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad album uri: %s", err), http.StatusBadRequest)
		return
	}

	album, err := spotify.GetAlbum(ctx, spClient, albumID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get album: %s", err), http.StatusInternalServerError)
		return
	}

	rows, err := history.AllAlbumStreamsByURI(
		ctx,
		db.Service().DB,
		userID,
		albumURI,
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
	response := AlbumStatsResponse{
		Streams:  streams,
		Album:    album,
		Rankings: []MonthRanking{},
	}

	filter := getFilterParams(r)
	allRankings, responseCode, err := history.AlbumStreamRankingsByMonth(ctx, db.Service().DB, userID, filter, 30)
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
				response.Rankings = append(response.Rankings, ranking)
			}
		}
	}

	json.NewEncoder(w).Encode(response)
}

func getFilterParams(r *http.Request) history.FilterParams {

	minMSPlayedParam := r.URL.Query().Get("minimum_milliseconds")
	minMSPlayed, err := strconv.Atoi(minMSPlayedParam)
	if err != nil {
		minMSPlayed = DEFAULT_MIN_MS_FILTER
	}

	includeSkippedParam := r.URL.Query().Get("include_skipped")
	includeSkipped := strings.EqualFold(includeSkippedParam, "true")

	maxTracksParam := r.URL.Query().Get("max_tracks")
	maxTracks, err := strconv.Atoi(maxTracksParam)
	if err != nil {
		maxTracks = 30
	}

	artistURIParam := r.URL.Query().Get("artist_uri")
	var artistURI *string
	if artistURIParam != "" {
		artistURI = &artistURIParam
	}

	albumURIParam := r.URL.Query().Get("album_uri")
	var albumURI *string
	if albumURIParam != "" {
		albumURI = &albumURIParam
	}

	return history.FilterParams{
		MinMSPlayed:    int32(minMSPlayed),
		IncludeSkipped: includeSkipped,
		MaxTracks:      int32(maxTracks),
		ArtistURI:      artistURI,
		AlbumURI:       albumURI,
	}
}
