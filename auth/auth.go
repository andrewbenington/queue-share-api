package auth

import (
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	Scopes = []string{
		spotifyauth.ScopeUserReadPrivate,
		spotifyauth.ScopeUserReadPlaybackState,
		spotifyauth.ScopeUserModifyPlaybackState,
	}
)
