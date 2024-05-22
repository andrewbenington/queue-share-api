package controller

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/engine"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
)

const (
	DEFAULT_MIN_MS_FILTER = 30000
	DEFAULT_LIMIT         = 50
)

type HistoryEntry struct {
	Timestamp       time.Time        `json:"timestamp"`
	TrackName       string           `json:"track_name"`
	AlbumName       string           `json:"album_name"`
	ArtistName      string           `json:"artist_name"`
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

func (c *StatsController) GetAllHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
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

	entries := []HistoryEntry{}
	for _, row := range rows {
		entry := HistoryEntry{
			Timestamp:       row.Timestamp,
			TrackName:       row.TrackName,
			SpotifyTrackUri: row.SpotifyTrackUri,
			AlbumName:       row.AlbumName,
			ArtistName:      row.ArtistName,
			SpotifyAlbumUri: row.SpotifyAlbumUri.String,
			MsPlayed:        row.MsPlayed,
			Artists: append(
				[]db.TrackArtist{{
					Name: row.ArtistName,
					URI:  row.SpotifyArtistUri.String,
					ID:   service.IDFromURIMust(row.SpotifyArtistUri.String),
				}},
				row.OtherArtists...),
			ImageURL: row.ImageUrl,
		}

		if len(row.OtherArtists) > 0 {
			entry.Artists = append(entry.Artists, row.OtherArtists...)
		}

		entries = append(entries, entry)
	}

	json.NewEncoder(w).Encode(HistoryResponse{
		History:     entries,
		LastFetched: engine.LastFetch,
	})
}

func (c *StatsController) UploadHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userUUIDFromRequest(r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
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

func (c *StatsController) GetTopTracksByYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	filter := getFilterParams(r)

	results, responseCode, err := history.TrackStreamCountByYear(ctx, db.Service().DB, userUUID, filter)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	json.NewEncoder(w).Encode(results)
}

func (c *StatsController) GetTopAlbumsByYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	filter := getFilterParams(r)

	results, responseCode, err := history.AlbumStreamCountByYear(ctx, db.Service().DB, userUUID, filter)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	json.NewEncoder(w).Encode(results)
}

func (c *StatsController) GetTopArtistsByYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	filter := getFilterParams(r)

	results, responseCode, err := history.ArtistStreamCountByYear(ctx, db.Service().DB, userUUID, filter)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	json.NewEncoder(w).Encode(results)
}

func (c *StatsController) GetAllStreamsByURI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	rows, err := history.AllTrackStreamsByURI(
		ctx,
		db.Service().DB,
		userUUID,
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
	Year                 int               `json:"year"`
	Month                int               `json:"month"`
	Position             int               `json:"position"`
	StartDateUnixSeconds int64             `json:"start_date_unix_seconds"`
	Timeframe            history.Timeframe `json:"timeframe"`
}

func getFilterParams(r *http.Request) history.FilterParams {

	minMSPlayedParam := r.URL.Query().Get("minimum_milliseconds")
	minMSPlayed, err := strconv.Atoi(minMSPlayedParam)
	if err != nil {
		minMSPlayed = DEFAULT_MIN_MS_FILTER
	}

	includeSkippedParam := r.URL.Query().Get("include_skipped")
	includeSkipped := strings.EqualFold(includeSkippedParam, "true")

	maxParam := r.URL.Query().Get("max")
	max, err := strconv.Atoi(maxParam)
	if err != nil {
		log.Printf("Bad max param: '%s'", maxParam)
		max = 30
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

	timeframeParam := r.URL.Query().Get("timeframe")
	var timeframe history.Timeframe = "month"
	if timeframeParam == "day" || timeframeParam == "week" || timeframeParam == "year" {
		timeframe = history.Timeframe(timeframeParam)
	}

	return history.FilterParams{
		MinMSPlayed:    int32(minMSPlayed),
		IncludeSkipped: includeSkipped,
		Max:            int32(max),
		ArtistURI:      artistURI,
		AlbumURI:       albumURI,
		Timeframe:      timeframe,
	}
}
