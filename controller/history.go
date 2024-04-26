package controller

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/db/gen"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/google/uuid"
	z_spotify "github.com/zmb3/spotify/v2"
)

const (
	DEFAULT_MIN_MS_FILTER = 30000
)

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

	includeSkippedParam := r.URL.Query().Get("include_skipped")
	includeSkipped := strings.EqualFold(includeSkippedParam, "true")

	rows, err := gen.New(db.Service().DB).HistoryGetAll(ctx, gen.HistoryGetAllParams{userUUID, int32(minMSPlayed), includeSkipped})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}
	json.NewEncoder(w).Encode(rows)
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

		err = history.InsertEntries(ctx, transaction, userUUID, entryData)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error uploading history", http.StatusInternalServerError)
			return
		}

		fmt.Printf("Uploaded file %s\n", file.Name)

		// for _, entry := range entryData {
		// 	dbRow, err := insertParamsFromStreamingEntry(entry, userUUID)
		// 	if err != nil {
		// 		fmt.Println(err)
		// 		continue
		// 	}

		// 	gen.New(transaction).HistoryInsertOne(ctx, dbRow)

		// }

	}
	err = transaction.Commit()
	if err != nil {
		http.Error(w, "Error committing DB transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func insertParamsFromStreamingEntry(entry history.StreamingEntry, userUUID uuid.UUID) (gen.HistoryInsertOneParams, error) {
	parsedTime, err := time.Parse("2006-01-02T15:04:05Z", entry.Timestamp)
	if err != nil {
		return gen.HistoryInsertOneParams{}, err
	}

	row := gen.HistoryInsertOneParams{
		UserID:          userUUID,
		Timestamp:       parsedTime,
		Platform:        entry.Platform,
		MsPlayed:        entry.MsPlayed,
		ConnCountry:     entry.ConnCountry,
		IpAddr:          NullStringFromPtr(entry.IpAddr),
		UserAgent:       NullStringFromPtr(entry.UserAgent),
		TrackName:       entry.TrackName,
		ArtistName:      entry.ArtistName,
		AlbumName:       entry.AlbumName,
		SpotifyTrackUri: entry.SpotifyTrackUri,
		ReasonStart:     NullStringFromPtr(entry.ReasonStart),
		ReasonEnd:       NullStringFromPtr(entry.ReasonEnd),
		Shuffle:         entry.Shuffle,
		Skipped:         NullBoolFromPtr(entry.Skipped),
		Offline:         entry.Offline,
		IncognitoMode:   entry.IncognitoMode,
	}
	if entry.OfflineTimestamp != 0 {
		row.OfflineTimestamp = sql.NullTime{Valid: true, Time: time.Unix(entry.OfflineTimestamp/1000, entry.OfflineTimestamp%1000)}
	}

	return row, nil
}

func NullStringFromPtr(ptr *string) sql.NullString {
	if ptr == nil {
		return sql.NullString{}
	}
	return sql.NullString{Valid: true, String: *ptr}
}

func NullBoolFromPtr(ptr *bool) sql.NullBool {
	if ptr == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Valid: true, Bool: *ptr}
}

type MonthTopSongs struct {
	Year  int             `json:"year"`
	Month int             `json:"month"`
	Songs SongRankingList `json:"songs"`
}

func (m *MonthTopSongs) print() {
	fmt.Printf("%d/%d:\n", m.Month+1, m.Year)
	for i, song := range m.Songs {
		if song == nil {
			fmt.Printf("%d. (nil)\n", i+1)
			continue
		}
		fmt.Printf("%d. %s (%d plays)\n", i+1, song.Track.Name, song.Plays)
	}
}

type SongRanking struct {
	Track *z_spotify.FullTrack `json:"track"`
	URI   string               `json:"spotify_uri"`
	Plays int                  `json:"play_count"`
}

type SongRankingList []*SongRanking

func (p SongRankingList) Len() int           { return len(p) }
func (p SongRankingList) Less(i, j int) bool { return p[i].Plays < p[j].Plays }
func (p SongRankingList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (c *Controller) GetTopSongsByMonth(w http.ResponseWriter, r *http.Request) {
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

	includeSkippedParam := r.URL.Query().Get("include_skipped")
	includeSkipped := strings.EqualFold(includeSkippedParam, "true")

	results := []*MonthTopSongs{}

	songCounts := map[string]int{}
	var lastTimestamp *time.Time

	rows, err := gen.New(db.Service().DB).HistoryGetAll(ctx, gen.HistoryGetAllParams{userUUID, int32(minMSPlayed), includeSkipped})
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	rowsByURI := map[string]*gen.HistoryGetAllRow{}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	for i, row := range rows {
		rowsByURI[row.SpotifyTrackUri] = row
		if lastTimestamp != nil && (row.Timestamp.Month() != lastTimestamp.Month() || i == len(rows)-1) {
			fmt.Printf("Ranking songs for %s\n", lastTimestamp.Format(time.DateOnly))
			rankedSongs, err := rankBySongCount(ctx, songCounts, rowsByURI, spClient)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if len(rankedSongs) > 10 {
				rankedSongs = rankedSongs[:10]
			}

			monthData := MonthTopSongs{
				Month: int(lastTimestamp.Month()),
				Year:  lastTimestamp.Year(),
				Songs: rankedSongs,
			}
			results = append(results, &monthData)
			songCounts = map[string]int{}
		}

		lastTimestamp = &row.Timestamp

		if _, ok := songCounts[row.SpotifyTrackUri]; !ok {
			songCounts[row.SpotifyTrackUri] = 1
		} else {
			songCounts[row.SpotifyTrackUri] += 1
		}
	}

	trackIDs := []string{}
	presentTrackIDs := map[string]bool{}
	for _, monthTopSongs := range results {
		for _, song := range monthTopSongs.Songs {
			id, err := spotifyIDFromURI(song.URI)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if _, ok := presentTrackIDs[id]; !ok {
				trackIDs = append(trackIDs, id)
				presentTrackIDs[id] = true
			}
		}
	}

	trackData := map[string]z_spotify.FullTrack{}
	for start := 0; start < len(trackIDs); start += 50 {
		end := start + 50
		if end > len(trackIDs) {
			end = len(trackIDs)
		}
		fmt.Printf("Getting tracks loop %d-%d\n", start, end)
		results, err := spotify.GetTracks(ctx, spClient, trackIDs[start:end])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for id, data := range results {
			trackData[id] = data
		}
	}

	for _, monthTopSongs := range results {
		for _, track := range monthTopSongs.Songs {
			id, err := spotifyIDFromURI(track.URI)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if fullTrack, ok := trackData[id]; ok {
				track.Track = &fullTrack
			}
		}
	}
	fmt.Println("Getting tracks loop done")

	json.NewEncoder(w).Encode(results)
}

func rankBySongCount(ctx context.Context, songCounts map[string]int, rowsByURI map[string]*gen.HistoryGetAllRow, spClient *z_spotify.Client) (SongRankingList, error) {
	pl := SongRankingList{}
	i := 0
	for uri, count := range songCounts {
		ranking := SongRanking{
			URI:   uri,
			Plays: count}
		pl = append(pl, &ranking)
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl, nil
}

func spotifyIDFromURI(uri string) (string, error) {
	segments := strings.Split(uri, ":")
	if len(segments) != 3 {
		return "", fmt.Errorf("bad uri format (%s)", uri)
	}
	return segments[2], nil
}

func (c *Controller) GetTopTracksByYear(w http.ResponseWriter, r *http.Request) {
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

	includeSkippedParam := r.URL.Query().Get("include_skipped")
	includeSkipped := strings.EqualFold(includeSkippedParam, "true")

	timestampRange, err := gen.New(db.Service().DB).HistoryGetTimestampRange(ctx, userUUID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}
	minYear := timestampRange.First.Year()
	maxYear := timestampRange.Last.Year()

	results := map[int][]*gen.HistoryGetTrackStreamsByYearRow{}

	for year := minYear; year <= maxYear; year++ {
		rows, err := gen.New(db.Service().DB).HistoryGetTrackStreamsByYear(ctx, gen.HistoryGetTrackStreamsByYearParams{
			UserID:       userUUID,
			MinMsPlayed:  int32(minMSPlayed),
			IncludeSkips: includeSkipped,
			Year:         int32(year),
		})
		if err != nil {
			requests.RespondWithDBError(w, err)
			return
		}
		results[year] = rows
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

	includeSkippedParam := r.URL.Query().Get("include_skipped")
	includeSkipped := strings.EqualFold(includeSkippedParam, "true")

	timestampRange, err := gen.New(db.Service().DB).HistoryGetTimestampRange(ctx, userUUID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}
	minYear := timestampRange.First.Year()
	maxYear := timestampRange.Last.Year()

	results := map[int][]*gen.HistoryGetAlbumStreamsByYearRow{}

	for year := minYear; year <= maxYear; year++ {
		rows, err := gen.New(db.Service().DB).HistoryGetAlbumStreamsByYear(ctx, gen.HistoryGetAlbumStreamsByYearParams{
			UserID:       userUUID,
			MinMsPlayed:  int32(minMSPlayed),
			IncludeSkips: includeSkipped,
			Year:         int32(year),
		})
		if err != nil {
			requests.RespondWithDBError(w, err)
			return
		}
		results[year] = rows
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

	includeSkippedParam := r.URL.Query().Get("include_skipped")
	includeSkipped := strings.EqualFold(includeSkippedParam, "true")

	timestampRange, err := gen.New(db.Service().DB).HistoryGetTimestampRange(ctx, userUUID)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}
	minYear := timestampRange.First.Year()
	maxYear := timestampRange.Last.Year()

	results := map[int][]*gen.HistoryGetArtistStreamsByYearRow{}

	for year := minYear; year <= maxYear; year++ {
		rows, err := gen.New(db.Service().DB).HistoryGetArtistStreamsByYear(ctx, gen.HistoryGetArtistStreamsByYearParams{
			UserID:       userUUID,
			MinMsPlayed:  int32(minMSPlayed),
			IncludeSkips: includeSkipped,
			Year:         int32(year),
		})
		if err != nil {
			requests.RespondWithDBError(w, err)
			return
		}
		results[year] = rows
	}

	json.NewEncoder(w).Encode(results)
}
