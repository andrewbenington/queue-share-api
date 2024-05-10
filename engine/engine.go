package engine

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	cycle_period                 = time.Second * 10
	cycle_period_history         = time.Minute * 30
	cycle_period_load_uris       = time.Hour * 3
	cycle_period_spotify_profile = time.Hour * 24
)

var (
	last_cycle_history         *time.Time
	last_cycle_load_uris       *time.Time
	last_cycle_spotify_profile *time.Time
)

func Run() {
	for {
		cycle()
		// if time.Now().Hour() < 6 {
		// fmt.Println("night; sleeping for 3 hours")
		// time.Sleep(CYCLE_PERIOD_NIGHT)
		// } else {
		log.Printf("engine sleeping for %s", cycle_period.String())

		time.Sleep(cycle_period)

		// }
	}
}

func shouldDoCycle(last *time.Time, period time.Duration) bool {
	return last == nil || time.Since(*last) >= period
}

func cycle() {
	fmt.Println("good morning: starting engine cycle")
	ctx := context.Background()
	now := time.Now()

	if strings.EqualFold(os.Getenv("IS_PROD"), "true") {
		if shouldDoCycle(last_cycle_history, cycle_period_history) {
			fmt.Println("doing history cycle")
			doHistoryCycle(ctx)
			last_cycle_history = &now
		}

		if shouldDoCycle(last_cycle_spotify_profile, cycle_period_spotify_profile) {
			fmt.Println("doing profile cycle")
			doSpotifyProfileCycle()
			last_cycle_spotify_profile = &now

			fmt.Println("uploading cache")
			uploadTrackCache(ctx)
			uploadAlbumCache(ctx)
			uploadArtistCache(ctx)
		}

		if shouldDoCycle(last_cycle_load_uris, cycle_period_load_uris) {
			fmt.Println("doing uri cycle")
			loadURIsByPopularity()
			last_cycle_load_uris = &now
		}
	}
	fmt.Println("doing track cache cycle")
	cacheTracksByPopularity()
	fmt.Println("doing album cache cycle")
	cacheAlbumsByPopularity()

	fmt.Println("engine cycle complete")
}
