package controller

import (
	"encoding/json"
)

type Controller struct{}

type ErrorResponse struct {
	Error string `json:"error"`
}

func MarshalErrorBody(e string) []byte {
	body, err := json.MarshalIndent(ErrorResponse{Error: e}, "", " ")
	if err != nil {
		body, _ = json.MarshalIndent(ErrorResponse{Error: err.Error()}, "", " ")
	}
	return body
}
