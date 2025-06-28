package history

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/zmb3/spotify/v2"
)

type Timeframe string

const (
	TimeframeDay     Timeframe = "day"
	TimeframeWeek    Timeframe = "week"
	TimeframeMonth   Timeframe = "month"
	TimeframeYear    Timeframe = "year"
	TimeframeAllTime Timeframe = "all_time"
)

func (t Timeframe) GetNextStartTime(current time.Time) time.Time {
	switch t {
	case TimeframeDay:
		return current.AddDate(0, 0, 1)
	case TimeframeWeek:
		return current.AddDate(0, 0, 7)
	case TimeframeMonth:
		return current.AddDate(0, 1, 0)
	case TimeframeYear:
		return current.AddDate(1, 0, 0)
	default:
		return time.Now()
	}
}

func (t Timeframe) DefaultFirstStartTime() *time.Time {
	now := time.Now()
	switch t {
	case TimeframeDay:
		monthAgo := now.AddDate(0, -1, 0)
		start := time.Date(monthAgo.Year(), monthAgo.Month(), monthAgo.Day(), 0, 0, 0, 0, time.Local)
		return &start
	case TimeframeWeek:
		twelveWeeksAgo := now.AddDate(0, 0, -12*7)
		start := time.Date(twelveWeeksAgo.Year(), twelveWeeksAgo.Month(), twelveWeeksAgo.Day(), 0, 0, 0, 0, time.Local)
		return &start
	default:
		return nil
	}
}

func (t Timeframe) GetEarliestStartTime(end time.Time) *time.Time {
	switch t {
	case TimeframeDay:
		monthBefore := end.AddDate(0, -1, 0)
		return &monthBefore
	case TimeframeWeek:
		twelveWeeksBefore := end.AddDate(0, 0, -12*7)
		return &twelveWeeksBefore
	default:
		return nil
	}
}

func GetSpotifyHistoryForUser(ctx context.Context, user *db.UserGetByIDRow) ([]spotify.RecentlyPlayedItem, error) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database to get missing artist URIs %w", err)
	}
	defer tx.Commit(ctx)

	_, spClient, err := client.ForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	token, _ := spClient.Token()
	log.Printf("%+v", token)

	var rows []spotify.RecentlyPlayedItem
	var before *time.Time

	for rows == nil || len(rows) > 0 {
		opts := spotify.RecentlyPlayedOptions{Limit: 50}
		if before != nil {
			opts.BeforeEpochMs = before.UnixMilli()
		}

		recentlyPlayed, err := spClient.PlayerRecentlyPlayedOpt(ctx, &opts)
		if err != nil {
			return nil, err
		} else {
			log.Printf("%d history entries found for %s", len(recentlyPlayed), user.Username)
		}
		if len(recentlyPlayed) == 0 {
			break
		}

		rows = append(rows, recentlyPlayed...)
		before = &recentlyPlayed[len(recentlyPlayed)-1].PlayedAt

	}

	return rows, nil
}
