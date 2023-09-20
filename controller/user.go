package controller

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewbenington/queue-share-api/constants"
	"github.com/andrewbenington/queue-share-api/db"
)

type createUserBody struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserBody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorBadRequest))
		return
	}

	userID, err := db.Service().UserStore.InsertUser(r.Context(), req.Name, req.Password)
	if err != nil {
		log.Printf("create user: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}

	respBody := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{
		ID:   userID,
		Name: req.Name,
	}
	bodyBytes, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("marshal create user response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(bodyBytes)
}
