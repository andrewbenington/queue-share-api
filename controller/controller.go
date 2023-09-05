package controller

import (
	"encoding/json"

	"github.com/zmb3/spotify/v2"
)

type Controller struct {
	Client     *spotify.Client
	Session    string
	NewSession string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewController(c *spotify.Client) *Controller {
	return &Controller{
		Client: c,
	}
}

func MarshalErrorBody(e error) []byte {
	body, err := json.MarshalIndent(ErrorResponse{Error: e.Error()}, "", " ")
	if err != nil {
		body, _ = json.MarshalIndent(ErrorResponse{Error: err.Error()}, "", " ")
	}
	return body
}
