package handlers

import (
	"net/http"
)

func DefaultHandler(env *Env, w http.ResponseWriter, r *http.Request) error {
	return StatusData{
		Code: http.StatusMethodNotAllowed,
		Data: map[string]string{
			"error": "Method is not allowed",
		},
	}
}
