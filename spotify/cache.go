package spotify

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/zmb3/spotify/v2"
)

var (
	spotifyTrackCache  = map[string]spotify.FullTrack{}
	spotifyArtistCache = map[string]spotify.FullArtist{}
	spotifyAlbumCache  = map[string]spotify.FullAlbum{}
)

func init() {
	// load track cache
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

	// load artist cache
	filePath = path.Join("temp", "artist_cache.json")
	if _, err := os.Stat(filePath); err != nil {
		return
	}
	bytes, err = os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error loading track data cache: %s\n", err)
		return
	}
	err = json.Unmarshal(bytes, &spotifyArtistCache)
	if err != nil {
		fmt.Printf("Error parsing artist data cache: %s\n", err)
	} else {
		fmt.Println("Artist cache loaded successfully")
	}

	// load album cache
	filePath = path.Join("temp", "album_cache.json")
	if _, err := os.Stat(filePath); err != nil {
		return
	}
	bytes, err = os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error loading track data cache: %s\n", err)
		return
	}
	err = json.Unmarshal(bytes, &spotifyAlbumCache)
	if err != nil {
		fmt.Printf("Error parsing album data cache: %s\n", err)
	} else {
		fmt.Println("Album cache loaded successfully")
	}
}

func cacheTracks(tracks []*spotify.FullTrack) {
	log.Println(tracks)
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

func cacheArtists(artists []*spotify.FullArtist) {
	for _, artist := range artists {
		spotifyArtistCache[string(artist.ID)] = *artist
	}
	bytes, err := json.Marshal(spotifyArtistCache)
	if err != nil {
		fmt.Printf("Error serializing Spotify artist cache: %s", err)
		return
	}
	err = os.WriteFile(path.Join("temp", "artist_cache.json"), bytes, 0644)
	if err != nil {
		fmt.Printf("Error serializing Spotify artist cache: %s", err)
	}
}

func getArtistsFromCache(ids []string) map[string]spotify.FullArtist {
	cacheHits := map[string]spotify.FullArtist{}
	for _, id := range ids {
		if artistData, ok := spotifyArtistCache[id]; ok {
			cacheHits[id] = artistData
		}
	}
	return cacheHits
}

func cacheAlbums(albums []*spotify.FullAlbum) {
	log.Println(albums)
	for _, album := range albums {
		spotifyAlbumCache[string(album.ID)] = *album
	}
	bytes, err := json.Marshal(spotifyAlbumCache)
	if err != nil {
		fmt.Printf("Error serializing Spotify album cache: %s", err)
		return
	}
	err = os.WriteFile(path.Join("temp", "album_cache.json"), bytes, 0644)
	if err != nil {
		fmt.Printf("Error serializing Spotify album cache: %s", err)
	}
}

func getAlbumsFromCache(ids []string) map[string]spotify.FullAlbum {
	cacheHits := map[string]spotify.FullAlbum{}
	for _, id := range ids {
		if albumData, ok := spotifyAlbumCache[id]; ok {
			cacheHits[id] = albumData
		}
	}
	return cacheHits
}

func GetTrackCache() map[string]spotify.FullTrack {
	return spotifyTrackCache
}

func GetAlbumCache() map[string]spotify.FullAlbum {
	return spotifyAlbumCache
}

func GetArtistCache() map[string]spotify.FullArtist {
	return spotifyArtistCache
}
