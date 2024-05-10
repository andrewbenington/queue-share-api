package history

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/google/uuid"
	z_spotify "github.com/zmb3/spotify/v2"
)

type StreamingEntry struct {
	Timestamp        string  `json:"ts"`
	Username         string  `json:"username"`
	Platform         string  `json:"platform"`
	MsPlayed         int32   `json:"ms_played"`
	ConnCountry      string  `json:"conn_country"`
	IpAddr           *string `json:"ip_addr_decrypted"`
	UserAgent        *string `json:"user_agent_decrypted"`
	TrackName        string  `json:"master_metadata_track_name"`
	ArtistName       string  `json:"master_metadata_album_artist_name"`
	AlbumName        string  `json:"master_metadata_album_album_name"`
	SpotifyTrackUri  string  `json:"spotify_track_uri"`
	ReasonStart      *string `json:"reason_start"`
	ReasonEnd        *string `json:"reason_end"`
	Shuffle          bool    `json:"shuffle"`
	Skipped          *bool   `json:"skipped"`
	Offline          bool    `json:"offline"`
	OfflineTimestamp int64   `json:"offline_timestamp"`
	IncognitoMode    bool    `json:"incognito_mode"`
}

func InsertEntry(ctx context.Context, transaction db.DBTX, entry db.HistoryInsertOneParams) error {
	return db.New(transaction).HistoryInsertOne(ctx, entry)
}

func InsertEntries(ctx context.Context, transaction db.DBTX, entries []db.HistoryInsertOneParams) error {
	params := db.HistoryInsertBulkParams{
		UserIds:          []uuid.UUID{},
		Timestamp:        []time.Time{},
		Platform:         []string{},
		MsPlayed:         []int32{},
		ConnCountry:      []string{},
		IpAddr:           []string{},
		UserAgent:        []string{},
		TrackName:        []string{},
		ArtistName:       []string{},
		AlbumName:        []string{},
		SpotifyTrackUri:  []string{},
		ReasonStart:      []string{},
		ReasonEnd:        []string{},
		Shuffle:          []bool{},
		Skipped:          []bool{},
		Offline:          []bool{},
		IncognitoMode:    []bool{},
		OfflineTimestamp: []time.Time{},
		FromHistory:      []bool{},
	}

	for _, entry := range entries {
		if entry.TrackName == "" {
			fmt.Println("skipping track with no name")
			continue
		}

		params.UserIds = append(params.UserIds, entry.UserID)
		params.Timestamp = append(params.Timestamp, entry.Timestamp)
		params.Platform = append(params.Platform, entry.Platform)
		params.MsPlayed = append(params.MsPlayed, entry.MsPlayed)
		params.ConnCountry = append(params.ConnCountry, entry.ConnCountry)
		params.IpAddr = append(params.IpAddr, entry.IpAddr.String)
		params.UserAgent = append(params.UserAgent, entry.UserAgent.String)
		params.TrackName = append(params.TrackName, entry.TrackName)
		params.ArtistName = append(params.ArtistName, entry.ArtistName)
		params.AlbumName = append(params.AlbumName, entry.AlbumName)
		params.SpotifyTrackUri = append(params.SpotifyTrackUri, entry.SpotifyTrackUri)
		params.SpotifyArtistUri = append(params.SpotifyArtistUri, entry.SpotifyArtistUri.String)
		params.SpotifyAlbumUri = append(params.SpotifyAlbumUri, entry.SpotifyAlbumUri.String)
		params.ReasonStart = append(params.ReasonStart, entry.ReasonStart.String)
		params.ReasonEnd = append(params.ReasonEnd, entry.ReasonEnd.String)
		params.Shuffle = append(params.Shuffle, entry.Shuffle)
		params.Skipped = append(params.Skipped, entry.Skipped.Bool)
		params.Offline = append(params.Offline, entry.Offline)
		params.IncognitoMode = append(params.IncognitoMode, entry.IncognitoMode)
		params.OfflineTimestamp = append(params.OfflineTimestamp, entry.OfflineTimestamp.Time)
		params.FromHistory = append(params.FromHistory, entry.FromHistory)
	}
	return db.New(transaction).HistoryInsertBulk(ctx, params)
}

func InsertEntriesFromHistory(ctx context.Context, transaction db.DBTX, userID uuid.UUID, entries []StreamingEntry) error {
	params := db.HistoryInsertBulkParams{
		UserIds:          []uuid.UUID{},
		Timestamp:        []time.Time{},
		Platform:         []string{},
		MsPlayed:         []int32{},
		ConnCountry:      []string{},
		IpAddr:           []string{},
		UserAgent:        []string{},
		TrackName:        []string{},
		ArtistName:       []string{},
		AlbumName:        []string{},
		SpotifyTrackUri:  []string{},
		ReasonStart:      []string{},
		ReasonEnd:        []string{},
		Shuffle:          []bool{},
		Skipped:          []bool{},
		Offline:          []bool{},
		IncognitoMode:    []bool{},
		OfflineTimestamp: []time.Time{},
		FromHistory:      []bool{},
	}

	for _, entry := range entries {
		if entry.TrackName == "" {
			continue
		}

		params.UserIds = append(params.UserIds, userID)

		parsedTime, err := time.Parse("2006-01-02T15:04:05Z", entry.Timestamp)
		if err != nil {
			fmt.Println(err)
			continue
		}
		params.Timestamp = append(params.Timestamp, parsedTime)
		params.Platform = append(params.Platform, entry.Platform)
		params.MsPlayed = append(params.MsPlayed, entry.MsPlayed)
		params.ConnCountry = append(params.ConnCountry, entry.ConnCountry)
		params.IpAddr = append(params.IpAddr, NullStringFromPtr(entry.IpAddr).String)
		params.UserAgent = append(params.UserAgent, NullStringFromPtr(entry.UserAgent).String)
		params.TrackName = append(params.TrackName, entry.TrackName)
		params.ArtistName = append(params.ArtistName, entry.ArtistName)
		params.AlbumName = append(params.AlbumName, entry.AlbumName)
		params.SpotifyTrackUri = append(params.SpotifyTrackUri, entry.SpotifyTrackUri)
		params.ReasonStart = append(params.ReasonStart, NullStringFromPtr(entry.ReasonStart).String)
		params.ReasonEnd = append(params.ReasonEnd, NullStringFromPtr(entry.ReasonEnd).String)
		params.Shuffle = append(params.Shuffle, entry.Shuffle)
		params.Skipped = append(params.Skipped, NullBoolFromPtr(entry.Skipped).Bool)
		params.Offline = append(params.Offline, entry.Offline)
		params.IncognitoMode = append(params.IncognitoMode, entry.IncognitoMode)
		params.OfflineTimestamp = append(params.OfflineTimestamp, time.Unix(entry.OfflineTimestamp/1000, entry.OfflineTimestamp%1000))
		params.FromHistory = append(params.FromHistory, true)
	}
	return db.New(transaction).HistoryInsertBulk(ctx, params)
}

type FilterParams struct {
	MinMSPlayed    int32
	IncludeSkipped bool
	MaxTracks      int32
	ArtistURI      *string
	AlbumURI       *string
}

func (f *FilterParams) ensureMinimum() {
	if f.MinMSPlayed == 0 {
		f.MinMSPlayed = 30000
	}
}

func FullHistoryTimeRange(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID) (minYear int, maxYear int, err error) {
	timestampRange, err := db.New(transaction).HistoryGetTimestampRange(ctx, userUUID)
	if err != nil {
		return 0, 0, err
	}

	return timestampRange.First.Year(), timestampRange.Last.Year(), nil
}

func AllTrackStreamsByURI(ctx context.Context, transaction db.DBTX, userID string, uri string, filter FilterParams) ([]*db.HistoryGetByTrackURIRow, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user UUID: %w", err)
	}

	filter.ensureMinimum()

	return db.New(transaction).HistoryGetByTrackURI(ctx, db.HistoryGetByTrackURIParams{
		UserID:       userUUID,
		MinMsPlayed:  filter.MinMSPlayed,
		IncludeSkips: filter.IncludeSkipped,
		URI:          uri,
	})
}

func AllArtistStreamsByURI(ctx context.Context, transaction db.DBTX, userID string, uri string, filter FilterParams) ([]*db.HistoryGetByArtistURIRow, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user UUID: %w", err)
	}

	filter.ensureMinimum()

	return db.New(transaction).HistoryGetByArtistURI(ctx, db.HistoryGetByArtistURIParams{
		UserID:       userUUID,
		MinMsPlayed:  filter.MinMSPlayed,
		IncludeSkips: filter.IncludeSkipped,
		URI:          sql.NullString{Valid: true, String: uri},
	})
}

func AllAlbumStreamsByURI(ctx context.Context, transaction db.DBTX, userID string, uri string, filter FilterParams) ([]*db.HistoryGetByAlbumURIRow, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user UUID: %w", err)
	}

	filter.ensureMinimum()

	return db.New(transaction).HistoryGetByAlbumURI(ctx, db.HistoryGetByAlbumURIParams{
		UserID:       userUUID,
		MinMsPlayed:  filter.MinMSPlayed,
		IncludeSkips: filter.IncludeSkipped,
		URI:          sql.NullString{Valid: true, String: uri},
	})
}

type StreamCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func TrackStreamCountByYear(ctx context.Context, transaction db.DBTX, userID string, filter FilterParams) (map[int][]StreamCount, int, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("parse user UUID: %w", err)
	}

	filter.ensureMinimum()

	minYear, maxYear, err := FullHistoryTimeRange(ctx, transaction, userUUID)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	results := map[int][]StreamCount{}

	for year := minYear; year <= maxYear; year++ {
		rows, err := db.New(db.Service().DB).HistoryGetTrackStreamCountByYear(ctx, db.HistoryGetTrackStreamCountByYearParams{
			UserID:       userUUID,
			MinMsPlayed:  filter.MinMSPlayed,
			IncludeSkips: filter.IncludeSkipped,
			Year:         int32(year),
		})

		if err != nil {
			fmt.Println("Error getting")
		}

		counts := []StreamCount{}
		for _, row := range rows {
			counts = append(counts, StreamCount{row.TrackName, row.Occurrences})
		}
		results[year] = counts
	}

	return results, http.StatusOK, nil
}

type MonthTopTracks struct {
	Year   int             `json:"year"`
	Month  int             `json:"month"`
	Tracks []*TrackStreams `json:"tracks"`
}

type TrackStreams struct {
	ID            string `json:"spotify_id"`
	Streams       int    `json:"stream_count"`
	StreamsChange *int64 `json:"streams_change,omitempty"`
	RankChange    *int64 `json:"rank_change,omitempty"`
}

func TrackStreamRankingsByMonth(ctx context.Context, transaction db.DBTX, userID string, filter FilterParams) ([]*MonthTopTracks, int, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("parse user UUID: %w", err)
	}

	minYear, maxYear, err := FullHistoryTimeRange(ctx, transaction, userUUID)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	results := []*MonthTopTracks{}

	lastMonthStreams := map[string]int64{}
	thisMonthStreams := map[string]int64{}

	lastMonthRanks := map[string]int64{}
	thisMonthRanks := map[string]int64{}

	for year := minYear; year <= maxYear; year++ {
		for month := 1; month <= 12; month++ {
			rows, err := db.New(db.Service().DB).HistoryGetTopTracksInTimeframe(ctx, db.HistoryGetTopTracksInTimeframeParams{
				UserID:       userUUID,
				MinMsPlayed:  filter.MinMSPlayed,
				IncludeSkips: filter.IncludeSkipped,
				StartDate:    time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local),
				EndDate:      time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.Local),
				MaxTracks:    filter.MaxTracks,
				ArtistURI:    NullStringFromPtr(filter.ArtistURI),
				AlbumURI:     NullStringFromPtr(filter.AlbumURI),
			})
			if err != nil {
				return nil, http.StatusNotFound, err
			}
			rankingList := []*TrackStreams{}
			for rankMinusOne, row := range rows {
				trackStreams := TrackStreams{
					ID:      spotify.IDFromURIMust(row.SpotifyTrackUri),
					Streams: int(row.Occurrences),
				}

				thisMonthStreams[row.SpotifyTrackUri] = row.Occurrences
				thisMonthRanks[row.SpotifyTrackUri] = int64(rankMinusOne) + 1

				if lastStreams, ok := lastMonthStreams[row.SpotifyTrackUri]; ok {
					diff := row.Occurrences - lastStreams
					trackStreams.StreamsChange = &diff
				}

				if lastRank, ok := lastMonthRanks[row.SpotifyTrackUri]; ok {
					diff := lastRank - (int64(rankMinusOne) + 1)
					trackStreams.RankChange = &diff
				}

				rankingList = append(rankingList, &trackStreams)
			}
			if len(rankingList) > 10 {
				rankingList = rankingList[:10]
			}

			results = append(results, &MonthTopTracks{
				Year:   year,
				Month:  month,
				Tracks: rankingList,
			})

			lastMonthStreams = thisMonthStreams
			thisMonthStreams = map[string]int64{}

			lastMonthRanks = thisMonthRanks
			thisMonthRanks = map[string]int64{}
		}
	}

	return results, 0, nil
}

type MonthTopArtists struct {
	Year    int              `json:"year"`
	Month   int              `json:"month"`
	Artists []*ArtistStreams `json:"artists"`
}

type ArtistStreams struct {
	ID            string                `json:"spotify_id"`
	Streams       int64                 `json:"stream_count"`
	StreamsChange *int64                `json:"streams_change,omitempty"`
	RankChange    *int64                `json:"rank_change,omitempty"`
	Artist        *z_spotify.FullArtist `json:"artist,omitempty"`
	Tracks        []string              `json:"tracks"`
}

func ArtistStreamRankingsByMonth(ctx context.Context, transaction db.DBTX, userID string, filter FilterParams, maxArtists int32) ([]*MonthTopArtists, int, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("parse user UUID: %w", err)
	}

	minYear, maxYear, err := FullHistoryTimeRange(ctx, transaction, userUUID)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	results := []*MonthTopArtists{}

	lastMonthStreams := map[string]int64{}
	thisMonthStreams := map[string]int64{}

	lastMonthRanks := map[string]int64{}
	thisMonthRanks := map[string]int64{}

	for year := minYear; year <= maxYear; year++ {
		for month := 1; month <= 12; month++ {
			rows, err := db.New(db.Service().DB).HistoryGetTopArtistsInTimeframe(ctx, db.HistoryGetTopArtistsInTimeframeParams{
				UserID:       userUUID,
				MinMsPlayed:  filter.MinMSPlayed,
				IncludeSkips: filter.IncludeSkipped,
				StartDate:    time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local),
				EndDate:      time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.Local),
				MaxTracks:    maxArtists,
			})

			if err != nil {
				return nil, http.StatusNotFound, err
			}

			rankingList := []*ArtistStreams{}
			for rankMinusOne, row := range rows {
				if !row.SpotifyArtistUri.Valid {
					continue
				}
				artistStreams := ArtistStreams{
					ID:      spotify.IDFromURIMust(row.SpotifyArtistUri.String),
					Streams: row.Occurrences,
					Tracks:  strings.Split(row.Tracks, "|~|"),
				}

				thisMonthStreams[row.SpotifyArtistUri.String] = row.Occurrences
				thisMonthRanks[row.SpotifyArtistUri.String] = int64(rankMinusOne) + 1

				if lastStreams, ok := lastMonthStreams[row.SpotifyArtistUri.String]; ok {
					diff := row.Occurrences - lastStreams
					artistStreams.StreamsChange = &diff
				}

				if lastRank, ok := lastMonthRanks[row.SpotifyArtistUri.String]; ok {
					diff := lastRank - (int64(rankMinusOne) + 1)
					artistStreams.RankChange = &diff
				}

				rankingList = append(rankingList, &artistStreams)
			}

			results = append(results, &MonthTopArtists{
				Year:    year,
				Month:   month,
				Artists: rankingList,
			})

			lastMonthStreams = thisMonthStreams
			thisMonthStreams = map[string]int64{}

			lastMonthRanks = thisMonthRanks
			thisMonthRanks = map[string]int64{}
		}
	}

	return results, 0, nil
}

type MonthTopAlbums struct {
	Year   int             `json:"year"`
	Month  int             `json:"month"`
	Albums []*AlbumStreams `json:"albums"`
}

type AlbumStreams struct {
	ID            string               `json:"spotify_id"`
	Streams       int                  `json:"stream_count"`
	StreamsChange *int64               `json:"streams_change,omitempty"`
	RankChange    *int64               `json:"rank_change,omitempty"`
	Album         *z_spotify.FullAlbum `json:"album,omitempty"`
	Tracks        []string             `json:"tracks"`
}

func AlbumStreamRankingsByMonth(ctx context.Context, transaction db.DBTX, userID string, filter FilterParams, maxAlbums int32) ([]*MonthTopAlbums, int, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("parse user UUID: %w", err)
	}

	minYear, maxYear, err := FullHistoryTimeRange(ctx, transaction, userUUID)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	results := []*MonthTopAlbums{}

	lastMonthStreams := map[string]int64{}
	thisMonthStreams := map[string]int64{}

	lastMonthRanks := map[string]int64{}
	thisMonthRanks := map[string]int64{}

	for year := minYear; year <= maxYear; year++ {
		for month := 1; month <= 12; month++ {
			rows, err := db.New(db.Service().DB).HistoryGetTopAlbumsInTimeframe(ctx, db.HistoryGetTopAlbumsInTimeframeParams{
				UserID:       userUUID,
				MinMsPlayed:  filter.MinMSPlayed,
				IncludeSkips: filter.IncludeSkipped,
				StartDate:    time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local),
				EndDate:      time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.Local),
				MaxTracks:    maxAlbums,
				ArtistURI:    NullStringFromPtr(filter.ArtistURI),
			})

			if err != nil {
				return nil, http.StatusNotFound, err
			}

			rankingList := []*AlbumStreams{}
			for rankMinusOne, row := range rows {
				if !row.SpotifyAlbumUri.Valid {
					continue
				}
				albumStreams := AlbumStreams{
					ID:      spotify.IDFromURIMust(row.SpotifyAlbumUri.String),
					Streams: int(row.Occurrences),
					Tracks:  strings.Split(row.Tracks, "|~|"),
				}

				thisMonthStreams[row.SpotifyAlbumUri.String] = row.Occurrences
				thisMonthRanks[row.SpotifyAlbumUri.String] = int64(rankMinusOne) + 1

				if lastStreams, ok := lastMonthStreams[row.SpotifyAlbumUri.String]; ok {
					diff := row.Occurrences - lastStreams
					albumStreams.StreamsChange = &diff
				}

				if lastRank, ok := lastMonthRanks[row.SpotifyAlbumUri.String]; ok {
					diff := lastRank - (int64(rankMinusOne) + 1)
					albumStreams.RankChange = &diff
				}

				rankingList = append(rankingList, &albumStreams)
			}

			results = append(results, &MonthTopAlbums{
				Year:   year,
				Month:  month,
				Albums: rankingList,
			})

			lastMonthStreams = thisMonthStreams
			thisMonthStreams = map[string]int64{}

			lastMonthRanks = thisMonthRanks
			thisMonthRanks = map[string]int64{}
		}
	}

	return results, 0, nil
}

func AlbumStreamCountByYear(ctx context.Context, transaction db.DBTX, userID string, filter FilterParams) (map[int][]StreamCount, int, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("parse user UUID: %w", err)
	}

	filter.ensureMinimum()

	minYear, maxYear, err := FullHistoryTimeRange(ctx, transaction, userUUID)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	results := map[int][]StreamCount{}

	for year := minYear; year <= maxYear; year++ {
		rows, err := db.New(db.Service().DB).HistoryGetAlbumStreamCountByYear(ctx, db.HistoryGetAlbumStreamCountByYearParams{
			UserID:       userUUID,
			MinMsPlayed:  filter.MinMSPlayed,
			IncludeSkips: filter.IncludeSkipped,
			Year:         int32(year),
		})

		if err != nil {
			fmt.Println("Error getting")
			continue
		}

		counts := []StreamCount{}
		for _, row := range rows {
			counts = append(counts, StreamCount{row.AlbumName, row.Occurrences})
		}
		results[year] = counts
	}

	return results, http.StatusOK, nil
}

func ArtistStreamCountByYear(ctx context.Context, transaction db.DBTX, userID string, filter FilterParams) (map[int][]StreamCount, int, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("parse user UUID: %w", err)
	}

	filter.ensureMinimum()

	minYear, maxYear, err := FullHistoryTimeRange(ctx, transaction, userUUID)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	results := map[int][]StreamCount{}

	for year := minYear; year <= maxYear; year++ {
		rows, err := db.New(db.Service().DB).HistoryGetArtistStreamCountByYear(ctx, db.HistoryGetArtistStreamCountByYearParams{
			UserID:       userUUID,
			MinMsPlayed:  filter.MinMSPlayed,
			IncludeSkips: filter.IncludeSkipped,
			Year:         int32(year),
		})

		if err != nil {
			fmt.Println("Error getting")
			continue
		}

		counts := []StreamCount{}
		for _, row := range rows {
			counts = append(counts, StreamCount{row.ArtistName, row.Occurrences})
		}
		results[year] = counts
	}

	return results, http.StatusOK, nil
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
