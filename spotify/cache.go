package spotify

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/zmb3/spotify/v2"
)

var (
	spotifyTrackCache = map[string]spotify.FullTrack{}
)

func init() {
	filePath := path.Join("temp", "track_cache.json")
	if _, err := os.Stat(filePath); err != nil {
		return
	}
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error loading track data cache: %s\n", err)
		return
	}
	err = json.Unmarshal(bytes, &spotifyTrackCache)
	if err != nil {
		fmt.Printf("Error parsing track data cache: %s\n", err)
	} else {
		fmt.Println("Track cache loaded successfully")
	}
}

func cacheTracks(tracks []*spotify.FullTrack) {
	for _, track := range tracks {
		spotifyTrackCache[string(track.ID)] = *track
	}
	bytes, err := json.Marshal(spotifyTrackCache)
	if err != nil {
		fmt.Printf("Error serializing Spotify track cache: %s", err)
		return
	}
	err = os.WriteFile(path.Join("temp", "track_cache.json"), bytes, 0644)
	if err != nil {
		fmt.Printf("Error serializing Spotify track cache: %s", err)
	}
}

func getTracksFromCache(ids []string) map[string]spotify.FullTrack {
	cacheHits := map[string]spotify.FullTrack{}
	for _, id := range ids {
		if trackData, ok := spotifyTrackCache[id]; ok {
			cacheHits[id] = trackData
		}
	}
	return cacheHits
}
