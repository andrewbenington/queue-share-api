package user

type User struct {
	ID           string `json:"user_id"`
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	SpotifyName  string `json:"spotify_name"`
	SpotifyImage string `json:"spotify_image"`
}
