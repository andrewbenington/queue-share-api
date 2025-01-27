package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/andrewbenington/queue-share-api/config"
	"github.com/andrewbenington/queue-share-api/util"
)

const (
	cycle_period_history         = time.Minute * 30
	cycle_period_spotify_profile = time.Hour * 24
	cycle_period_save_logs       = time.Second * 10
)

var (
	cycle_period               = time.Second * 10
	cycle_period_load_uris     = time.Minute * 30
	last_cycle_history         *time.Time
	last_cycle_load_uris       *time.Time
	last_cycle_spotify_profile *time.Time
	last_cycle_save_logs       *time.Time
)

func Run() {
	for {
		cycle()
		// log.Printf("engine sleeping for %s", cycle_period.String())

		time.Sleep(cycle_period)

	}
}

func shouldDoCycle(last *time.Time, period time.Duration) bool {
	return last == nil || time.Since(*last) >= period
}

func cycle() {
	// fmt.Println("good morning: starting engine cycle")
	ctx := context.Background()
	now := time.Now()

	if config.GetIsProd() {
		if shouldDoCycle(last_cycle_load_uris, cycle_period_load_uris) {
			fmt.Println("doing uri cycle")
			total := loadURIsByPopularity(ctx)
			fmt.Println("Total:", total)
			if total == 0 {
				cycle_period_load_uris = time.Minute * 30
			} else {
				cycle_period_load_uris = time.Second * 10
			}
			last_cycle_load_uris = &now
		}

		if shouldDoCycle(last_cycle_history, cycle_period_history) {
			fmt.Println("doing history cycle")
			doHistoryCycle(ctx)
			last_cycle_history = &now
		}

		if shouldDoCycle(last_cycle_spotify_profile, cycle_period_spotify_profile) {
			fmt.Println("doing profile cycle")
			doSpotifyProfileCycle(ctx)
			last_cycle_spotify_profile = &now

			fmt.Println("uploading cache")
			uploadAlbumCache(ctx)
		}
	}

	if shouldDoCycle(last_cycle_save_logs, cycle_period_save_logs) {
		// fmt.Println("doing log cycle")
		util.WriteChannelLogsToFile()
		last_cycle_save_logs = &now
	}

	// fmt.Println("engine cycle complete")
}
