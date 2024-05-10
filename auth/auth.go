package auth

import (
	"sync"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type SpotifyLoginState struct {
	UserID      string
	RedirectURI string
}

type UserContextKeyT struct{}

var (
	SpotifyScopes = []string{
		spotifyauth.ScopeUserReadPrivate,
		spotifyauth.ScopeUserReadPlaybackState,
		spotifyauth.ScopeUserModifyPlaybackState,
		spotifyauth.ScopeUserTopRead,
		spotifyauth.ScopeUserReadRecentlyPlayed,
	}
	SpotifyStates     = map[string]SpotifyLoginState{}
	SpotifyStatesLock = sync.Mutex{}
	UserContextKey    UserContextKeyT
)
