package history

import (
	"context"
	"database/sql"
	"time"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/exp/maps"
)

type Node[T any] struct {
	Data T
	Prev *Node[T]
	Next *Node[T]
}

func (n *Node[T]) InsertAfter(data T) *Node[T] {
	newNode := Node[T]{
		Data: data,
		Next: n.Next,
		Prev: n,
	}

	if n.Next != nil {
		n.Next.Prev = &newNode
	}

	n.Next = &newNode

	return &newNode
}

func (n *Node[T]) InsertBefore(data T) *Node[T] {
	newNode := Node[T]{
		Data: data,
		Next: n,
		Prev: n.Prev,
	}

	if n.Prev != nil {
		n.Prev.Next = &newNode
	}

	n.Prev = &newNode

	return &newNode
}

func (this *Node[T]) SwapWithNext() {
	if this.Next == nil {
		return
	}

	afterThis := this.Next
	beforeThis := this.Prev
	twoAfterThis := afterThis.Next

	afterThis.Prev = beforeThis
	afterThis.Next = this
	this.Prev = afterThis
	this.Next = twoAfterThis

	if twoAfterThis != nil {
		twoAfterThis.Prev = this
	}

	if beforeThis != nil {
		beforeThis.Next = afterThis
	}

}

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

	if len(rankings) == 0 {
		return nil, nil
	}

	nodesByISRC := map[string]*Node[TrackStreams]{}
	trackIDs := map[string]bool{}

	linkedList := &Node[TrackStreams]{Data: *rankings[0]}
	nodesByISRC[rankings[0].ISRC] = linkedList
	trackIDs[rankings[0].ID] = true

	for _, rank := range rankings[1:] {
		linkedList = linkedList.InsertBefore(*rank)
		nodesByISRC[rank.ISRC] = linkedList
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
		StartDate:    sql.NullTime{Valid: true, Time: *eventsStart},
		EndDate:      sql.NullTime{Valid: true, Time: *eventsEnd},
		MaxCount:     20000,
	})

	rankEvents := []TrackRankEvent{}

	for _, stream := range lo.Reverse(allStreams) {
		rankEvent := TrackRankEvent{}
		node, ok := nodesByISRC[stream.Isrc.String]
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

func GetArtistRankEvents(ctx context.Context, transaction db.DBTX, userUUID uuid.UUID, filter FilterParams) ([]ArtistRankEvent, error) {
	filter.ensureStartAndEnd()
	filter.Max = 50
	filter.IncludeSkipped = false

	_, _, rankings, err := CalcArtistStreamsAndRanks(ctx, userUUID, filter, transaction, time.Unix(0, 0), *filter.Start, nil, nil)
	if err != nil {
		return nil, err
	}

	if len(rankings) == 0 {
		return nil, nil
	}

	nodesByID := map[string]*Node[ArtistStreams]{}
	artistIDs := map[string]bool{}

	linkedList := &Node[ArtistStreams]{Data: *rankings[0]}
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

	allStreams, err := db.New(transaction).HistoryGetAll(ctx, db.HistoryGetAllParams{
		UserID:       userUUID,
		MinMsPlayed:  filter.MinMSPlayed,
		IncludeSkips: filter.IncludeSkipped,
		StartDate:    sql.NullTime{Valid: true, Time: *filter.Start},
		EndDate:      sql.NullTime{Valid: true, Time: *filter.End},
		MaxCount:     20000,
	})

	rankEvents := []ArtistRankEvent{}

	for _, stream := range lo.Reverse(allStreams) {
		rankEvent := ArtistRankEvent{}
		node, ok := nodesByID[service.IDFromURIMust(stream.SpotifyArtistUri.String)]
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

	nodesByID := map[string]*Node[AlbumStreams]{}
	albumIDs := map[string]bool{}

	linkedList := &Node[AlbumStreams]{Data: *rankings[0]}
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
		StartDate:    sql.NullTime{Valid: true, Time: *filter.Start},
		EndDate:      sql.NullTime{Valid: true, Time: *filter.End},
		MaxCount:     20000,
	})

	rankEvents := []AlbumRankEvent{}

	for _, stream := range lo.Reverse(allStreams) {
		rankEvent := AlbumRankEvent{}
		node, ok := nodesByID[service.IDFromURIMust(stream.SpotifyAlbumUri.String)]
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
