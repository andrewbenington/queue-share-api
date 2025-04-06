package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/andrewbenington/queue-share-api/util"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
)

type MixBuilder struct {
	ArtistIDs   []string `json:"artist_ids"`
	AlbumIDs    []string `json:"album_ids"`
	PlaylistIDs []string `json:"playlist_ids"`
}

func (c *StatsController) AddMixToQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		fmt.Println(err)
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	var req MixBuilder
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println(err)
		requests.RespondBadRequest(w)
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	playlistTracks := []spotify.FullTrack{}

	for _, playlistID := range req.PlaylistIDs {
		playlist, err := spClient.GetPlaylist(ctx, spotify.ID(playlistID))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, track := range playlist.Tracks.Tracks {
			if strings.HasPrefix(string(track.Track.URI), "spotify:local") {
				continue
			}
			playlistTracks = append(playlistTracks, track.Track)
		}
	}

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}
	defer tx.Rollback(ctx)

	albums, err := service.GetAlbums(ctx, spClient, req.AlbumIDs)
	if err != nil {
		fmt.Println(err, "could not get albums")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	artistURIs := lo.Map(
		req.ArtistIDs,
		func(id string, _ int) string {
			if id == "" {
				fmt.Printf("MISSING ARTIST")
			}
			return fmt.Sprintf("spotify:artist:%s", id)
		})

	artistStreams, err := db.New(tx).HistoryGetRecentArtistStreams(ctx, db.HistoryGetRecentArtistStreamsParams{
		UserID:     userUUID,
		ArtistUris: artistURIs,
	})
	if err != nil {
		fmt.Println(err, "could not get artist streams")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shuffledTrackURIs := shuffleAndSeparateSameArtist(artistStreams, albums, playlistTracks)

	uriCount := min(len(shuffledTrackURIs), 60)
	for _, uri := range shuffledTrackURIs[:uriCount] {
		err = service.PushToUserQueue(r.Context(), spClient, service.IDFromURIMust(uri))
		if err != nil && strings.Contains(err.Error(), "No active device found") {
			requests.RespondWithError(w, http.StatusBadRequest, "Host is not playing music")
			return
		}
		if err != nil {
			fmt.Println(err, "could not enqueue")
			requests.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	tx.Commit(ctx)
	w.WriteHeader(http.StatusAccepted)
}

type TrackToShuffle interface {
	artistURI() string
	trackURI() string
	source() string
}

type ArtistTrackToShuffle struct {
	row *db.HistoryGetRecentArtistStreamsRow
}

func (t *ArtistTrackToShuffle) artistURI() string {
	return *t.row.SpotifyArtistUri
}

func (t *ArtistTrackToShuffle) trackURI() string {
	return t.row.SpotifyTrackUri
}

func (t *ArtistTrackToShuffle) source() string {
	return "artist"
}

type AlbumTrackToShuffle struct {
	spotifyTrackURI string
	album           db.AlbumData
}

func (t *AlbumTrackToShuffle) artistURI() string {
	return t.album.URI
}

func (t *AlbumTrackToShuffle) trackURI() string {
	return t.spotifyTrackURI
}

func (t *AlbumTrackToShuffle) source() string {
	return "album"
}

type SpotifyTrackToShuffle struct {
	track spotify.FullTrack
}

func (t *SpotifyTrackToShuffle) artistURI() string {
	return string(t.track.Artists[0].URI)
}

func (t *SpotifyTrackToShuffle) trackURI() string {
	return string(t.track.URI)
}

func (t *SpotifyTrackToShuffle) source() string {
	return "track"
}

func shuffleAndSeparateSameArtist(fromArtists []*db.HistoryGetRecentArtistStreamsRow, fromAlbums map[string]db.AlbumData, fromPlaylists []spotify.FullTrack) []string {

	toShuffle := []TrackToShuffle{}
	included := map[string]struct{}{}

	for _, row := range fromArtists {
		if _, ok := included[row.SpotifyTrackUri]; ok {
			continue
		}
		toShuffle = append(toShuffle, &ArtistTrackToShuffle{
			row: row,
		})
		included[row.SpotifyTrackUri] = struct{}{}
	}

	for _, album := range fromAlbums {
		for _, trackID := range album.SpotifyTrackIds {
			uri := fmt.Sprintf("spotify:track:%s", trackID)
			if _, ok := included[uri]; ok {
				continue
			}
			toShuffle = append(toShuffle, &AlbumTrackToShuffle{
				spotifyTrackURI: uri,
				album:           album,
			})
			included[uri] = struct{}{}
		}
	}

	for _, track := range fromPlaylists {
		if _, ok := included[string(track.URI)]; ok {
			continue
		}
		toShuffle = append(toShuffle, &SpotifyTrackToShuffle{
			track: track,
		})
		included[string(track.URI)] = struct{}{}
	}

	toShuffle = shuffleAndSeparateByArtist(toShuffle)
	uris := make([]string, 0, len(toShuffle))
	for _, track := range toShuffle {
		uris = append(uris, track.trackURI())
	}

	return uris
}

type TracksToShuffle = []TrackToShuffle

func shuffleAndSeparateByArtist(tracks []TrackToShuffle) []TrackToShuffle {

	perm := rand.Perm(len(tracks))
	shuffled := util.DoublyLinkedList[TrackToShuffle]{}
	for _, v := range perm {
		shuffled.PushEnd(tracks[v])
	}

	shuffledAndSeparated := util.DoublyLinkedList[TrackToShuffle]{}
	prevSize := 0
	log.Printf("Starting separation (size %d)...", shuffled.Size())

	toAdd := &shuffled
	for prevSize != toAdd.Size() {
		log.Printf("Separation pass (%d separated, %d remaining)...", shuffledAndSeparated.Size(), toAdd.Size())
		prevSize = toAdd.Size()
		toAddNext := util.DoublyLinkedList[TrackToShuffle]{}
		for toAdd.Size() > 0 {
			log.Printf("toAdd size: %d", toAdd.Size())
			next := toAdd.PopFirst()
			if last := shuffledAndSeparated.PeekLast(); last == nil || (*next).artistURI() != (*last).artistURI() {
				log.Printf("Pushing artist %s", (*next).artistURI())
				shuffledAndSeparated.PushEnd(*next)
			} else {
				log.Printf("Artist %s previously pushed, moving %s to the end", (*next).artistURI(), (*next).trackURI())
				toAddNext.PushEnd(*next)
			}
		}
		toAdd = &toAddNext
	}

	return shuffledAndSeparated.ToSlice()
}
