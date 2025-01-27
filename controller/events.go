package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/samber/lo"
)

func (c *StatsController) GetRecentUserEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	filter := getFilterParams(r)

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	trackEvents, err := history.GetTrackRankEvents(ctx, tx, userUUID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	artistEvents, err := history.GetArtistRankEvents(ctx, tx, userUUID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	albumEvents, err := history.GetAlbumRankEvents(ctx, tx, userUUID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	allEvents := []history.RankEvent{}

	iTrack, iArtist, iAlbum := 0, 0, 0

	for iTrack < len(trackEvents) || iArtist < len(artistEvents) || iAlbum < len(albumEvents) {
		if trackEventIsNext(trackEvents, iTrack, artistEvents, iArtist, albumEvents, iAlbum) {
			allEvents = append(allEvents, &trackEvents[iTrack])
			iTrack++
		} else if artistEventIsNext(trackEvents, iTrack, artistEvents, iArtist, albumEvents, iAlbum) {
			allEvents = append(allEvents, &artistEvents[iArtist])
			iArtist++
		} else {
			allEvents = append(allEvents, &albumEvents[iAlbum])
			iAlbum++
		}
	}

	json.NewEncoder(w).Encode(allEvents)
}

func trackEventIsNext(
	trackEvents []history.TrackRankEvent, iTrack int,
	artistEvents []history.ArtistRankEvent, iArtist int,
	albumEvents []history.AlbumRankEvent, iAlbum int) bool {
	if iTrack >= len(trackEvents) {
		return false
	}

	if iArtist >= len(artistEvents) {
		if iAlbum >= len(albumEvents) {
			return true
		}
		return !trackEvents[iTrack].GetTime().After(albumEvents[iAlbum].GetTime())
	}

	if iAlbum >= len(albumEvents) {
		if iArtist >= len(artistEvents) {
			return true
		}
		return !trackEvents[iTrack].GetTime().After(artistEvents[iArtist].GetTime())
	}

	return trackEvents[iTrack].GetTime().Before(artistEvents[iArtist].GetTime()) && trackEvents[iTrack].GetTime().Before(albumEvents[iAlbum].GetTime())
}

func artistEventIsNext(
	trackEvents []history.TrackRankEvent, iTrack int,
	artistEvents []history.ArtistRankEvent, iArtist int,
	albumEvents []history.AlbumRankEvent, iAlbum int) bool {
	if iArtist >= len(artistEvents) {
		return false
	}

	if iTrack >= len(trackEvents) {
		if iAlbum >= len(albumEvents) {
			return true
		}
		return artistEvents[iArtist].GetTime().Before(albumEvents[iAlbum].GetTime())
	}

	if iAlbum >= len(albumEvents) {
		if iTrack >= len(trackEvents) {
			return true
		}
		return artistEvents[iArtist].GetTime().Before(trackEvents[iTrack].GetTime())
	}

	return artistEvents[iArtist].GetTime().Before(albumEvents[iAlbum].GetTime()) && artistEvents[iArtist].GetTime().Before(trackEvents[iTrack].GetTime())
}

type NewArtistsResponseEntry struct {
	StreamCount int           `json:"stream_count"`
	Artist      db.ArtistData `json:"artist"`
	FirstStream time.Time     `json:"first_stream"`
}

func (c *StatsController) GetNewArtists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	filter := getFilterParams(r)
	if filter.Start.Unix() == 0 {
		today := time.Now()
		monthAgo := today.Add(-30 * 24 * time.Hour)
		filter.Start = &monthAgo
		filter.End = &today
	}

	transaction, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer transaction.Commit(ctx)

	newArtistData, err := db.New(transaction).HistoryGetNewArtists(ctx, db.HistoryGetNewArtistsParams{
		UserID:    userUUID,
		StartDate: *filter.Start,
		EndDate:   *filter.End,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	ids := lo.Map(newArtistData, func(artist *db.HistoryGetNewArtistsRow, _ int) string { return artist.ID })
	artistMap, err := service.GetArtists(ctx, spClient, ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	response := []NewArtistsResponseEntry{}
	for _, newArtistDatum := range newArtistData {
		if artist, ok := artistMap[newArtistDatum.ID]; ok {
			firstStream, err := time.Parse("2006-01-02", newArtistDatum.DistinctDates[0])
			if err != nil {
				log.Println(err)
				continue
			}
			response = append(response, NewArtistsResponseEntry{
				StreamCount: int(newArtistDatum.Count),
				Artist:      artist,
				FirstStream: firstStream,
			})
		}
	}

	json.NewEncoder(w).Encode(response)
}
