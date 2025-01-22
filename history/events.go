package history

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/andrewbenington/queue-share-api/util"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/exp/maps"
)

type RankEvent interface {
	GetTime() time.Time
}

type TrackRankEvent struct {
	Track     *db.TrackData    `json:"track"`
	Rank      int64            `json:"rank"`
	Streams   int              `json:"streams"`
	Surpassed []TrackRankEvent `json:"surpassed"`
	DateUnix  int64            `json:"date_unix"`
}

func (e *TrackRankEvent) GetTime() time.Time {
	return time.Unix(e.DateUnix, 0)
}

type ArtistRankEvent struct {
	Artist    *spotify.FullArtist `json:"artist"`
	Rank      int64               `json:"rank"`
	Streams   int64               `json:"streams"`
	Surpassed []ArtistRankEvent   `json:"surpassed"`
	DateUnix  int64               `json:"date_unix"`
}

func (e *ArtistRankEvent) GetTime() time.Time {
	return time.Unix(e.DateUnix, 0)
}

type AlbumRankEvent struct {
	Album     *spotify.FullAlbum `json:"album"`
	Rank      int64              `json:"rank"`
	Streams   int64              `json:"streams"`
	Surpassed []AlbumRankEvent   `json:"surpassed"`
	DateUnix  int64              `json:"date_unix"`
}

func (e *AlbumRankEvent) GetTime() time.Time {
	return time.Unix(e.DateUnix, 0)
}

func GetTrackRankEvents(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, filter FilterParams) ([]TrackRankEvent, error) {
	filter.ensureStartAndEnd()
	eventsStart := filter.Start
	eventsEnd := filter.End
	filter.Max = 50
	filter.IncludeSkipped = false
	filter.End = filter.Start
	filter.Start = &time.Time{}

	_, _, rankings, err := CalcTrackStreamsAndRanks(ctx, userUUID, filter, transaction, nil, nil)
	if err != nil {
		return nil, err
	}

	if len(rankings) == 0 || rankings[0].ISRC == nil {
		return nil, nil
	}

	nodesByISRC := map[string]*util.Node[TrackStreams]{}
	trackIDs := map[string]bool{}

	linkedList := &util.Node[TrackStreams]{Data: *rankings[0]}
	nodesByISRC[*rankings[0].ISRC] = linkedList
	trackIDs[rankings[0].ID] = true

	for _, rank := range rankings[1:] {
		if rank.ISRC == nil {
			continue
		}
		linkedList = linkedList.InsertBefore(*rank)
		nodesByISRC[*rank.ISRC] = linkedList
		trackIDs[rank.ID] = true
	}

	_, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	tracksByID, err := service.GetTracks(ctx, spClient, maps.Keys(trackIDs))
	if err != nil {
		return nil, err
	}

	for linkedList != nil {
		linkedList = linkedList.Next
	}

	allStreams, err := db.New(transaction).HistoryGetAll(ctx, db.HistoryGetAllParams{
		UserID:       userUUID,
		MinMsPlayed:  filter.MinMSPlayed,
		IncludeSkips: filter.IncludeSkipped,
		StartDate:    eventsStart,
		EndDate:      eventsEnd,
		MaxCount:     20000,
	})

	rankEvents := []TrackRankEvent{}

	for _, stream := range lo.Reverse(allStreams) {
		if stream.Isrc == nil {
			continue
		}

		rankEvent := TrackRankEvent{}
		node, ok := nodesByISRC[*stream.Isrc]
		if !ok {
			continue
		}

		node.Data.Streams++

		for node.Next != nil && node.Next.Data.Streams < node.Data.Streams {
			track, ok := tracksByID[node.Data.ID]
			if !ok {
				continue
			}

			rankEvent.Track = &track
			rankEvent.Streams = node.Data.Streams
			rankEvent.DateUnix = stream.Timestamp.Unix()

			surpassedTrack, ok := tracksByID[node.Next.Data.ID]
			if ok {
				rankEvent.Surpassed = append(rankEvent.Surpassed, TrackRankEvent{
					Track:   &surpassedTrack,
					Rank:    node.Next.Data.Rank + 1,
					Streams: node.Next.Data.Streams,
				})
			}

			node.Data.Rank = node.Next.Data.Rank
			node.Next.Data.Rank += 1
			node.SwapWithNext()
			rankEvent.Rank = node.Data.Rank
			rankEvent.Streams = node.Data.Streams
		}
		if rankEvent.Track != nil {
			rankEvents = append(rankEvents, rankEvent)
		}
	}

	return rankEvents, err
}

func GetArtistRankEvents(ctx context.Context, tx db.DBTX, userUUID uuid.UUID, filter FilterParams) ([]ArtistRankEvent, error) {
	filter.ensureStartAndEnd()
	filter.Max = 50
	filter.IncludeSkipped = false

	_, _, rankings, err := CalcArtistStreamsAndRanks(ctx, userUUID, filter, tx, time.Unix(0, 0), *filter.Start, nil, nil)
	if err != nil {
		return nil, err
	}

	if len(rankings) == 0 {
		return nil, nil
	}

	nodesByID := map[string]*util.Node[ArtistStreams]{}
	artistIDs := map[string]bool{}

	linkedList := &util.Node[ArtistStreams]{Data: *rankings[0]}
	nodesByID[rankings[0].ID] = linkedList
	artistIDs[rankings[0].ID] = true

	for _, rank := range rankings[1:] {
		linkedList = linkedList.InsertBefore(*rank)
		nodesByID[rank.ID] = linkedList
		artistIDs[rank.ID] = true
	}

	_, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	artistsByID, err := service.GetArtists(ctx, spClient, maps.Keys(artistIDs))
	if err != nil {
		return nil, err
	}

	for linkedList != nil {
		linkedList = linkedList.Next
	}

	allStreams, err := db.New(tx).HistoryGetAll(ctx, db.HistoryGetAllParams{
		UserID:       userUUID,
		MinMsPlayed:  filter.MinMSPlayed,
		IncludeSkips: filter.IncludeSkipped,
		StartDate:    filter.Start,
		EndDate:      filter.End,
		MaxCount:     20000,
	})

	rankEvents := []ArtistRankEvent{}

	for _, stream := range lo.Reverse(allStreams) {
		if stream.SpotifyArtistUri == nil {
			log.Printf("no artist uri; skipping")
			continue
		}

		rankEvent := ArtistRankEvent{}
		id, err := service.IDFromURI(*stream.SpotifyArtistUri)
		if err != nil {
			return nil, fmt.Errorf("bad spotify uri: %s", *stream.SpotifyArtistUri)
		}

		node, ok := nodesByID[id]
		if !ok {
			continue
		}

		node.Data.Streams++

		for node.Next != nil && node.Next.Data.Streams < node.Data.Streams {
			artist, ok := artistsByID[node.Data.ID]
			if !ok {
				continue
			}

			rankEvent.Artist = &artist
			rankEvent.DateUnix = stream.Timestamp.Unix()
			rankEvent.Streams = node.Data.Streams

			surpassedArtist, ok := artistsByID[node.Next.Data.ID]
			if ok {
				rankEvent.Surpassed = append(rankEvent.Surpassed, ArtistRankEvent{
					Artist:  &surpassedArtist,
					Rank:    node.Next.Data.Rank + 1,
					Streams: node.Next.Data.Streams,
				})
			}
			node.Data.Rank = node.Next.Data.Rank
			node.Next.Data.Rank += 1
			node.SwapWithNext()
			rankEvent.Rank = node.Data.Rank
		}

		if rankEvent.Artist != nil {
			rankEvents = append(rankEvents, rankEvent)
		}
	}

	return rankEvents, err
}

func GetAlbumRankEvents(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, filter FilterParams) ([]AlbumRankEvent, error) {
	filter.ensureStartAndEnd()
	filter.Max = 50
	filter.IncludeSkipped = false

	_, _, rankings, err := CalcAlbumStreamsAndRanks(ctx, userUUID, filter, transaction, time.Unix(0, 0), *filter.Start, nil, nil)
	if err != nil {
		return nil, err
	}

	if len(rankings) == 0 {
		return nil, nil
	}

	nodesByID := map[string]*util.Node[AlbumStreams]{}
	albumIDs := map[string]bool{}

	linkedList := &util.Node[AlbumStreams]{Data: *rankings[0]}
	nodesByID[rankings[0].ID] = linkedList
	albumIDs[rankings[0].ID] = true

	for _, rank := range rankings[1:] {
		linkedList = linkedList.InsertBefore(*rank)
		nodesByID[rank.ID] = linkedList
		albumIDs[rank.ID] = true
	}

	_, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	albumsByID, err := service.GetAlbums(ctx, spClient, maps.Keys(albumIDs))
	if err != nil {
		return nil, err
	}

	for linkedList != nil {
		linkedList = linkedList.Next
	}

	allStreams, err := db.New(transaction).HistoryGetAll(ctx, db.HistoryGetAllParams{
		UserID:       userUUID,
		MinMsPlayed:  filter.MinMSPlayed,
		IncludeSkips: filter.IncludeSkipped,
		StartDate:    filter.Start,
		EndDate:      filter.End,
		MaxCount:     20000,
	})

	rankEvents := []AlbumRankEvent{}

	for _, stream := range lo.Reverse(allStreams) {
		if stream.SpotifyAlbumUri == nil {
			continue
		}

		rankEvent := AlbumRankEvent{}
		id, err := service.IDFromURI(*stream.SpotifyAlbumUri)
		if err != nil {
			return nil, fmt.Errorf("bad spotify uri: %s", *stream.SpotifyAlbumUri)
		}

		node, ok := nodesByID[id]
		if !ok {
			continue
		}

		node.Data.Streams++

		for node.Next != nil && node.Next.Data.Streams < node.Data.Streams {
			album, ok := albumsByID[node.Data.ID]
			if !ok {
				continue
			}

			rankEvent.Album = &album
			rankEvent.DateUnix = stream.Timestamp.Unix()
			rankEvent.Streams = node.Data.Streams

			surpassedAlbum, ok := albumsByID[node.Next.Data.ID]
			if ok {
				rankEvent.Surpassed = append(rankEvent.Surpassed, AlbumRankEvent{
					Album:   &surpassedAlbum,
					Rank:    node.Next.Data.Rank + 1,
					Streams: node.Next.Data.Streams,
				})
			}
			node.Data.Rank = node.Next.Data.Rank
			node.Next.Data.Rank += 1
			node.SwapWithNext()
			rankEvent.Rank = node.Data.Rank
		}

		if rankEvent.Album != nil {
			rankEvents = append(rankEvents, rankEvent)
		}
	}

	return rankEvents, err
}
