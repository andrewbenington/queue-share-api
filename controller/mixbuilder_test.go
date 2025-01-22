package controller

import (
	"testing"
)

type TestTrackToShuffle struct {
	artist string
	track  string
}

func (t TestTrackToShuffle) artistURI() string {
	return t.artist
}

func (t TestTrackToShuffle) trackURI() string {
	return t.track
}

func (t TestTrackToShuffle) source() string {
	return "test"
}

var (
	testData = []TestTrackToShuffle{
		{
			artist: "Lana Del Rey",
			track:  "Ride",
		},
		{
			artist: "Lana Del Rey",
			track:  "Sweet",
		},
		{
			artist: "Lana Del Rey",
			track:  "Say Yes to Heaven",
		},
		{
			artist: "Lana Del Rey",
			track:  "the greatest",
		},
		{
			artist: "Lana Del Rey",
			track:  "How to Disappear",
		},
		{
			artist: "Lana Del Rey",
			track:  "Kintsugi",
		},
		{
			artist: "Kylie Minogue",
			track:  "Magic",
		},
		{
			artist: "Kylie Minogue",
			track:  "Can't Get You Out Of My Head",
		},
		{
			artist: "Kylie Minogue",
			track:  "Fever",
		},
		{
			artist: "Kylie Minogue",
			track:  "Padam Padam",
		},
		{
			artist: "Kylie Minogue",
			track:  "Kiss Bang Bang",
		},
		{
			artist: "Kylie Minogue",
			track:  "Things We Do For Love",
		},
		{
			artist: "Magdalena Bay",
			track:  "The Beginning",
		},
		{
			artist: "Magdalena Bay",
			track:  "Cry For Me",
		},
		{
			artist: "Magdalena Bay",
			track:  "Top Dog",
		},
		{
			artist: "Magdalena Bay",
			track:  "Killing Time",
		},
		{
			artist: "Magdalena Bay",
			track:  "Secrets (Your Fire)",
		},
		{
			artist: "Magdalena Bay",
			track:  "Angel On A Satellite",
		},
		{
			artist: "Charli xcx",
			track:  "360",
		},
		{
			artist: "Charli xcx",
			track:  "Vroom Vroom",
		},
		{
			artist: "Charli xcx",
			track:  "Every Rule",
		},
		{
			artist: "Donna Summer",
			track:  "Heaven Knows",
		},
		{
			artist: "Donna Summer",
			track:  "Bad Girls",
		},
		{
			artist: "Donna Summer",
			track:  "Hot Stuff",
		},
		{
			artist: "Electric Light Orchestra",
			track:  "Tightrope",
		},
		{
			artist: "Electric Light Orchestra",
			track:  "Telephone Line",
		},
		{
			artist: "Electric Light Orchestra",
			track:  "Confusion",
		},
		{
			artist: "Electric Light Orchestra",
			track:  "Sweet Talkin' Woman",
		},
	}
)

func TestShuffleAndSeparateByArtist(t *testing.T) {
	interfaces := []TrackToShuffle{}
	for _, data := range testData {
		interfaces = append(interfaces, data)
	}
	for range 100 {
		shuffledAndSeparated := shuffleAndSeparateByArtist(interfaces)
		prevTrack := shuffledAndSeparated[0]
		for _, track := range shuffledAndSeparated[1:] {
			if track.artistURI() == prevTrack.artistURI() {
				t.Fatalf(`%s and %s should not be next to each other`, track.artistURI(), prevTrack.artistURI())
			}
			prevTrack = track
		}
	}
}
