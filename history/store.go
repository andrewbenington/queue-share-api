package history

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
)

var (
	trackRankingsCache  map[string][]*db.HistoryGetTopTracksInTimeframeDedupRow = make(map[string][]*db.HistoryGetTopTracksInTimeframeDedupRow)
	albumRankingsCache  map[string][]*db.HistoryGetTopAlbumsInTimeframeRow      = make(map[string][]*db.HistoryGetTopAlbumsInTimeframeRow)
	artistRankingsCache map[string][]*db.HistoryGetTopArtistsInTimeframeRow     = make(map[string][]*db.HistoryGetTopArtistsInTimeframeRow)

	trackCacheLock  sync.Mutex
	albumCacheLock  sync.Mutex
	artistCacheLock sync.Mutex
)

func init() {
	// load track rankings cache
	filePath := path.Join("temp", "track_rank_cache.json")
	if _, err := os.Stat(filePath); err != nil {
		return
	}
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error loading track rank cache: %s\n", err)
		return
	}
	err = json.Unmarshal(bytes, &trackRankingsCache)
	if err != nil {
		fmt.Printf("Error parsing track rank cache: %s\n", err)
	} else {
		fmt.Println("Track rank cache loaded successfully")
	}

	// load artist rankings cache
	filePath = path.Join("temp", "artist_rank_cache.json")
	if _, err := os.Stat(filePath); err != nil {
		return
	}
	bytes, err = os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error loading artist rank cache: %s\n", err)
		return
	}
	err = json.Unmarshal(bytes, &artistRankingsCache)
	if err != nil {
		fmt.Printf("Error parsing artist rank cache: %s\n", err)
	} else {
		fmt.Println("Artist rank cache loaded successfully")
	}
}

func WriteCachesToFile() {
	trackCacheLock.Lock()
	bytes, err := json.Marshal(trackRankingsCache)
	if err != nil {
		fmt.Printf("Error serializing track ranking cache: %s", err)
		return
	}
	err = os.WriteFile(path.Join("temp", "track_rank_cache.json"), bytes, 0644)
	if err != nil {
		fmt.Printf("Error serializing track ranking cache: %s", err)
	}
	trackCacheLock.Unlock()

	artistCacheLock.Lock()
	bytes, err = json.Marshal(artistRankingsCache)
	if err != nil {
		fmt.Printf("Error serializing artist ranking cache: %s", err)
		return
	}
	err = os.WriteFile(path.Join("temp", "artist_rank_cache.json"), bytes, 0644)
	if err != nil {
		fmt.Printf("Error serializing artist ranking cache: %s", err)
	}
	artistCacheLock.Unlock()
}

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
	trackIDs := lo.Map(entries, func(entry StreamingEntry, _ int) string {
		return service.IDFromURIMust(entry.SpotifyTrackUri)
	})
	cachedTracks, err := service.GetTracksFromCache(ctx, trackIDs)
	if err != nil {
		log.Printf("error getting tracks from cache: %s", err)
	}

	params := db.HistoryInsertBulkNullableParams{
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
		SpotifyArtistUri: []*string{},
		SpotifyAlbumUri:  []*string{},
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
		if cachedTrack, ok := cachedTracks[service.IDFromURIMust(entry.SpotifyTrackUri)]; ok {
			params.SpotifyAlbumUri = append(params.SpotifyAlbumUri, &cachedTrack.AlbumURI)
			params.SpotifyArtistUri = append(params.SpotifyArtistUri, &cachedTrack.ArtistURI)
		} else {
			params.SpotifyAlbumUri = append(params.SpotifyAlbumUri, nil)
			params.SpotifyArtistUri = append(params.SpotifyArtistUri, nil)
		}
		params.ReasonStart = append(params.ReasonStart, NullStringFromPtr(entry.ReasonStart).String)
		params.ReasonEnd = append(params.ReasonEnd, NullStringFromPtr(entry.ReasonEnd).String)
		params.Shuffle = append(params.Shuffle, entry.Shuffle)
		params.Skipped = append(params.Skipped, NullBoolFromPtr(entry.Skipped).Bool)
		params.Offline = append(params.Offline, entry.Offline)
		params.IncognitoMode = append(params.IncognitoMode, entry.IncognitoMode)
		params.OfflineTimestamp = append(params.OfflineTimestamp, time.Unix(entry.OfflineTimestamp/1000, entry.OfflineTimestamp%1000))
		params.FromHistory = append(params.FromHistory, true)
	}
	return db.New(transaction).HistoryInsertBulkNullable(ctx, params)
}

type FilterParams struct {
	MinMSPlayed    int32
	IncludeSkipped bool
	Max            int32
	ArtistURIs     []string
	AlbumURI       *string
	Timeframe      Timeframe
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

func AllTrackStreamsByURI(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, uri string, filter FilterParams) ([]*db.HistoryGetByTrackURIRow, error) {
	filter.ensureMinimum()

	return db.New(transaction).HistoryGetByTrackURI(ctx, db.HistoryGetByTrackURIParams{
		UserID:       userUUID,
		MinMsPlayed:  filter.MinMSPlayed,
		IncludeSkips: filter.IncludeSkipped,
		URI:          uri,
	})
}

func AllArtistStreamsByURI(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, uri string, filter FilterParams) ([]*db.HistoryGetByArtistURIRow, error) {
	filter.ensureMinimum()

	return db.New(transaction).HistoryGetByArtistURI(ctx, db.HistoryGetByArtistURIParams{
		UserID:       userUUID,
		MinMsPlayed:  filter.MinMSPlayed,
		IncludeSkips: filter.IncludeSkipped,
		URI:          sql.NullString{Valid: true, String: uri},
	})
}

func AllAlbumStreamsByURI(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, uri string, filter FilterParams) ([]*db.HistoryGetByAlbumURIRow, error) {
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

func TrackStreamCountByYear(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, filter FilterParams) (map[int][]StreamCount, int, error) {
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

func CalcTrackStreamsAndRanks(ctx context.Context, userUUID uuid.UUID, filter FilterParams, start time.Time, end time.Time, lastStreams map[string]int64, lastRanks map[string]int64) (
	streamsByURI map[string]int64,
	ranksByURI map[string]int64,
	rankingList []*TrackStreams,
	err error,
) {
	streamsByURI = map[string]int64{}
	ranksByURI = map[string]int64{}

	var rows []*db.HistoryGetTopTracksInTimeframeDedupRow

	cacheIdentifier := fmt.Sprintf("%s-%d-%d-%d", userUUID, start.UnixMilli(), end.UnixMilli(), filter.Max)
	if filter.AlbumURI != nil {
		cacheIdentifier += "-" + *filter.AlbumURI
	}
	if filter.ArtistURIs != nil {
		cacheIdentifier += "-" + strings.Join(filter.ArtistURIs, ",")
	}

	trackCacheLock.Lock()
	cachedRows, ok := trackRankingsCache[cacheIdentifier]
	trackCacheLock.Unlock()

	if ok && time.Since(end) >= time.Hour*24 {
		rows = cachedRows
	} else {
		rows, err = db.New(db.Service().DB).HistoryGetTopTracksInTimeframeDedup(ctx, db.HistoryGetTopTracksInTimeframeDedupParams{
			UserID:       userUUID,
			MinMsPlayed:  filter.MinMSPlayed,
			IncludeSkips: filter.IncludeSkipped,
			StartDate:    start.UTC(),
			EndDate:      end.UTC(),
			MaxTracks:    filter.Max + 20,
			ArtistUris:   filter.ArtistURIs,
			AlbumURI:     NullStringFromPtr(filter.AlbumURI),
		})
		if err != nil {
			return nil, nil, nil, err
		}
		trackCacheLock.Lock()
		trackRankingsCache[cacheIdentifier] = rows
		trackCacheLock.Unlock()
	}

	var prevCount int64 = 0
	var currentRank int64 = 0
	for _, row := range rows {
		if row.Occurrences != prevCount {
			currentRank++
			prevCount = row.Occurrences
		}

		trackStreams := TrackStreams{
			ID:      service.IDFromURIMust(row.SpotifyTrackUri),
			Streams: int(row.Occurrences),
			Rank:    currentRank,
		}

		streamsByURI[row.SpotifyTrackUri] = row.Occurrences
		ranksByURI[row.SpotifyTrackUri] = currentRank

		if lastStreams != nil {
			if lastStreams, ok := lastStreams[row.SpotifyTrackUri]; ok {
				diff := row.Occurrences - lastStreams
				trackStreams.StreamsChange = &diff
			}

		}

		if lastRanks != nil {
			if lastRank, ok := lastRanks[row.SpotifyTrackUri]; ok {
				diff := lastRank - currentRank
				trackStreams.RankChange = &diff
			}
		}

		rankingList = append(rankingList, &trackStreams)
	}

	if len(rankingList) > int(filter.Max) {
		rankingList = rankingList[:filter.Max]
	}

	return
}

func CalcAlbumStreamsAndRanks(ctx context.Context, userUUID uuid.UUID, filter FilterParams, start time.Time, end time.Time, lastStreams map[string]int64, lastRanks map[string]int64) (
	streamsByURI map[string]int64,
	ranksByURI map[string]int64,
	rankingList []*AlbumStreams,
	err error,
) {
	streamsByURI = map[string]int64{}
	ranksByURI = map[string]int64{}

	var rows []*db.HistoryGetTopAlbumsInTimeframeRow

	cacheIdentifier := fmt.Sprintf("%s-%d-%d-%d", userUUID, start.UnixMilli(), end.UnixMilli(), filter.Max)
	albumCacheLock.Lock()
	cachedRows, ok := albumRankingsCache[cacheIdentifier]
	albumCacheLock.Unlock()

	if ok && time.Since(end) >= time.Hour*24 {
		rows = cachedRows
	} else {
		rows, err = db.New(db.Service().DB).HistoryGetTopAlbumsInTimeframe(ctx, db.HistoryGetTopAlbumsInTimeframeParams{
			UserID:       userUUID,
			MinMsPlayed:  filter.MinMSPlayed,
			IncludeSkips: filter.IncludeSkipped,
			StartDate:    start.UTC(),
			EndDate:      end.UTC(),
			Max:          filter.Max + 20,
		})
		if err != nil {
			return nil, nil, nil, err
		}

		albumCacheLock.Lock()
		albumRankingsCache[cacheIdentifier] = rows
		albumCacheLock.Unlock()
	}

	var prevCount int64 = 0
	var currentRank int64 = 0
	for _, row := range rows {
		if row.Occurrences != prevCount {
			currentRank++
			prevCount = row.Occurrences
		}

		albumStreams := AlbumStreams{
			ID:      service.IDFromURIMust(row.SpotifyAlbumUri.String),
			Streams: row.Occurrences,
			Tracks:  strings.Split(row.Tracks, "|~|"),
			Rank:    currentRank,
		}

		streamsByURI[row.SpotifyAlbumUri.String] = row.Occurrences
		ranksByURI[row.SpotifyAlbumUri.String] = currentRank

		if lastStreams != nil {
			if lastStreams, ok := lastStreams[row.SpotifyAlbumUri.String]; ok {
				diff := row.Occurrences - lastStreams
				albumStreams.StreamsChange = &diff
			}

		}

		if lastRanks != nil {
			if lastRank, ok := lastRanks[row.SpotifyAlbumUri.String]; ok {
				diff := lastRank - currentRank
				albumStreams.RankChange = &diff
			}
		}

		rankingList = append(rankingList, &albumStreams)
	}
	if len(rankingList) > int(filter.Max) {
		rankingList = rankingList[:filter.Max]
	}

	return
}
func CalcArtistStreamsAndRanks(ctx context.Context, userUUID uuid.UUID, filter FilterParams, start time.Time, end time.Time, lastStreams map[string]int64, lastRanks map[string]int64) (
	streamsByURI map[string]int64,
	ranksByURI map[string]int64,
	rankingList []*ArtistStreams,
	err error,
) {
	streamsByURI = map[string]int64{}
	ranksByURI = map[string]int64{}

	var rows []*db.HistoryGetTopArtistsInTimeframeRow

	cacheIdentifier := fmt.Sprintf("%s-%d-%d-%d", userUUID, start.UnixMilli(), end.UnixMilli(), filter.Max)
	artistCacheLock.Lock()
	cachedRows, ok := artistRankingsCache[cacheIdentifier]
	artistCacheLock.Unlock()

	if ok && time.Since(end) >= time.Hour*24 {
		rows = cachedRows
	} else {
		rows, err = db.New(db.Service().DB).HistoryGetTopArtistsInTimeframe(ctx, db.HistoryGetTopArtistsInTimeframeParams{
			UserID:       userUUID,
			MinMsPlayed:  filter.MinMSPlayed,
			IncludeSkips: filter.IncludeSkipped,
			StartDate:    start.UTC(),
			EndDate:      end.UTC(),
			Max:          filter.Max + 20,
		})
		if err != nil {
			return nil, nil, nil, err
		}

		artistCacheLock.Lock()
		artistRankingsCache[cacheIdentifier] = rows
		artistCacheLock.Unlock()
	}

	var prevCount int64 = 0
	var currentRank int64 = 0
	for _, row := range rows {
		if row.Occurrences != prevCount {
			currentRank++
			prevCount = row.Occurrences
		}

		artistStreams := ArtistStreams{
			ID:      service.IDFromURIMust(row.SpotifyArtistUri.String),
			Streams: row.Occurrences,
			Rank:    currentRank,
			Tracks:  strings.Split(row.Tracks, "|~|"),
		}

		streamsByURI[row.SpotifyArtistUri.String] = row.Occurrences
		ranksByURI[row.SpotifyArtistUri.String] = currentRank

		if lastStreams != nil {
			if lastStreams, ok := lastStreams[row.SpotifyArtistUri.String]; ok {
				diff := row.Occurrences - lastStreams
				artistStreams.StreamsChange = &diff
			}

		}

		if lastRanks != nil {
			if lastRank, ok := lastRanks[row.SpotifyArtistUri.String]; ok {
				diff := lastRank - currentRank
				artistStreams.RankChange = &diff
			}
		}

		rankingList = append(rankingList, &artistStreams)
	}

	if len(rankingList) > int(filter.Max) {
		rankingList = rankingList[:filter.Max]
	}

	return
}

type TrackRankings struct {
	Tracks               []*TrackStreams `json:"tracks,omitempty"`
	StartDateUnixSeconds int64           `json:"start_date_unix_seconds"`
	Timeframe            Timeframe       `json:"timeframe"`
}

type TrackStreams struct {
	ID            string `json:"spotify_id"`
	Streams       int    `json:"stream_count"`
	StreamsChange *int64 `json:"streams_change,omitempty"`
	Rank          int64  `json:"rank"`
	RankChange    *int64 `json:"rank_change,omitempty"`
}

func TrackStreamRankingsByTimeframe(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, filter FilterParams, start *time.Time, end *time.Time) ([]*TrackRankings, int, error) {
	var firstStart time.Time
	endTime := time.Now()
	if end != nil {
		endTime = *end
	}

	if start != nil {
		firstStart = *start
	} else if defaultFirstStart := filter.Timeframe.DefaultFirstStartTime(); defaultFirstStart != nil {
		firstStart = *defaultFirstStart
	} else {
		minYear, _, err := FullHistoryTimeRange(ctx, transaction, userUUID)
		if err != nil {
			return nil, http.StatusNotFound, err
		}
		firstStart = time.Date(minYear, 0, 1, 0, 0, 0, 0, time.Local)
	}

	results := []*TrackRankings{}

	lastMonthStreams := map[string]int64{}
	lastMonthRanks := map[string]int64{}

	current := firstStart
	for current.Before(endTime) {
		nextStart := filter.Timeframe.GetNextStartTime(current)

		var rankingList []*TrackStreams
		thisMonthStreams, thisMonthRanks, rankingList, err := CalcTrackStreamsAndRanks(ctx, userUUID, filter, current, nextStart, lastMonthStreams, lastMonthRanks)
		if err != nil {
			return nil, http.StatusNotFound, err
		}

		results = append(results, &TrackRankings{
			Tracks:               rankingList,
			StartDateUnixSeconds: current.Unix(),
			Timeframe:            filter.Timeframe,
		})

		lastMonthStreams = thisMonthStreams
		lastMonthRanks = thisMonthRanks

		current = nextStart
	}

	return results, 0, nil
}

type ArtistRankings struct {
	Artists              []*ArtistStreams `json:"artists"`
	StartDateUnixSeconds int64            `json:"start_date_unix_seconds"`
	Timeframe            Timeframe        `json:"timeframe"`
}

type ArtistStreams struct {
	ID            string              `json:"spotify_id"`
	Streams       int64               `json:"stream_count"`
	StreamsChange *int64              `json:"streams_change,omitempty"`
	Rank          int64               `json:"rank"`
	RankChange    *int64              `json:"rank_change,omitempty"`
	Artist        *spotify.FullArtist `json:"artist,omitempty"`
	Tracks        []string            `json:"tracks"`
}

func ArtistStreamRankingsByTimeframe(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, filter FilterParams, start *time.Time, end *time.Time) ([]*ArtistRankings, int, error) {
	var firstStart time.Time
	endTime := time.Now()
	if end != nil {
		endTime = *end
	}

	if start != nil {
		firstStart = *start
	} else if defaultFirstStart := filter.Timeframe.DefaultFirstStartTime(); defaultFirstStart != nil {
		firstStart = *defaultFirstStart
	} else {
		minYear, _, err := FullHistoryTimeRange(ctx, transaction, userUUID)
		if err != nil {
			return nil, http.StatusNotFound, err
		}
		firstStart = time.Date(minYear, 0, 1, 0, 0, 0, 0, time.Local)
	}

	results := []*ArtistRankings{}

	lastMonthStreams := map[string]int64{}
	lastMonthRanks := map[string]int64{}

	current := firstStart
	for current.Before(endTime) {
		nextStart := filter.Timeframe.GetNextStartTime(current)

		var rankingList []*ArtistStreams
		thisMonthStreams, thisMonthRanks, rankingList, err := CalcArtistStreamsAndRanks(ctx, userUUID, filter, current, nextStart, lastMonthStreams, lastMonthRanks)
		if err != nil {
			return nil, http.StatusNotFound, err
		}

		results = append(results, &ArtistRankings{
			Artists:              rankingList,
			StartDateUnixSeconds: current.Unix(),
			Timeframe:            filter.Timeframe,
		})

		lastMonthStreams = thisMonthStreams
		lastMonthRanks = thisMonthRanks

		current = nextStart
	}

	return results, 0, nil
}

type AlbumRankings struct {
	AlbumStreams         []*AlbumStreams `json:"albums"`
	StartDateUnixSeconds int64           `json:"start_date_unix_seconds"`
	Timeframe            Timeframe       `json:"timeframe"`
}

type AlbumStreams struct {
	ID            string             `json:"spotify_id"`
	Streams       int64              `json:"stream_count"`
	StreamsChange *int64             `json:"streams_change,omitempty"`
	Rank          int64              `json:"rank"`
	RankChange    *int64             `json:"rank_change,omitempty"`
	Album         *spotify.FullAlbum `json:"album,omitempty"`
	Tracks        []string           `json:"tracks"`
}

func AlbumStreamRankingsByTimeframe(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, filter FilterParams, start *time.Time, end *time.Time) ([]*AlbumRankings, int, error) {

	var firstStart time.Time
	endTime := time.Now()
	if end != nil {
		endTime = *end
	}

	if start != nil {
		firstStart = *start
	} else if defaultFirstStart := filter.Timeframe.DefaultFirstStartTime(); defaultFirstStart != nil {
		firstStart = *defaultFirstStart
	} else {
		minYear, _, err := FullHistoryTimeRange(ctx, transaction, userUUID)
		if err != nil {
			return nil, http.StatusNotFound, err
		}
		firstStart = time.Date(minYear, 0, 1, 0, 0, 0, 0, time.Local)
	}

	results := []*AlbumRankings{}

	lastMonthStreams := map[string]int64{}
	lastMonthRanks := map[string]int64{}

	current := firstStart
	for current.Before(endTime) {
		nextStart := filter.Timeframe.GetNextStartTime(current)

		var rankingList []*AlbumStreams
		thisMonthStreams, thisMonthRanks, rankingList, err := CalcAlbumStreamsAndRanks(ctx, userUUID, filter, current, nextStart, lastMonthStreams, lastMonthRanks)
		if err != nil {
			return nil, http.StatusNotFound, err
		}

		results = append(results, &AlbumRankings{
			AlbumStreams:         rankingList,
			StartDateUnixSeconds: current.Unix(),
			Timeframe:            filter.Timeframe,
		})

		lastMonthStreams = thisMonthStreams
		lastMonthRanks = thisMonthRanks

		current = nextStart
	}

	return results, 0, nil
}

func AlbumStreamCountByYear(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, filter FilterParams) (map[int][]StreamCount, int, error) {
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

func ArtistStreamCountByYear(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, filter FilterParams) (map[int][]StreamCount, int, error) {
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
