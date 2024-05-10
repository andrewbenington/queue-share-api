package db

type TrackArtist struct {
	ID   string `json:"id"`
	URI  string `json:"uri"`
	Name string `json:"string"`
}

type TrackExternalIDs map[string]string
