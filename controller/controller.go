package controller

import "net/http"

type Controller struct{}

type StatsController struct{}

type StatsControllerInterface interface {
	GetTopAlbumsByMonth(w http.ResponseWriter, r *http.Request)
	GetTopArtistsByMonth(w http.ResponseWriter, r *http.Request)
	GetTopTracksByMonth(w http.ResponseWriter, r *http.Request)
}
