package user

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	SpotifyName  string `json:"spotify_name"`
	SpotifyImage string `json:"spotify_image"`
}
