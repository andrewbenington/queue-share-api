package controller

import (
	"encoding/json"
	"net/http"

	"github.com/andrewbenington/queue-share-api/version"
)

func (c *Controller) GetVersion(w http.ResponseWriter, r *http.Request) {
	v := version.Get()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}
