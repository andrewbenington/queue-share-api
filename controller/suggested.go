package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/spotify"
)

func (c *Controller) SuggestedTracks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Guest {
		requests.RespondAuthError(w)
		return
	}

	status, client, err := client.ForRoom(ctx, reqCtx.Room.Code)
	if err != nil {
		requests.RespondWithError(w, status, err.Error())
		return
	}

	tracks, err := spotify.TopTracks(ctx, client)
	if err != nil {
		log.Println(err)
		if strings.Contains(err.Error(), "403") {
			requests.RespondWithError(w, http.StatusForbidden, "Outdated token")
			return
		}
		requests.RespondInternalError(w)
		return
	}

	json.NewEncoder(w).Encode(tracks)
}
