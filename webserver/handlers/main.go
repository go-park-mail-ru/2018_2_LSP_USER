package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// StatusData represents an error with an associated HTTP status code.
type StatusData struct {
	Code int
	data interface{}
}

// Allows StatusData to satisfy the error interface.
func (sd StatusData) Error() string {
	return fmt.Sprintf("%v", sd.data)
}

// Allows StatusData to satisfy the error interface.
func (sd StatusData) GetJsonData() ([]byte, error) {
	return json.Marshal(sd.data)
}

// Returns our HTTP status code.
func (se StatusData) Status() int {
	return se.Code
}

// A (simple) example of our application-wide configuration.
type Env struct {
	DB       *sql.DB
	Host     string
	Database string
	Username string
	Password string
}

// HandlerFunc func for Handler
type HandlerFunc func(e *Env, w http.ResponseWriter, r *http.Request) error

type HandlersMap map[string]HandlerFunc

// The Handler struct that takes a configured Env and a function matching
// our useful signature.
type Handler struct {
	*Env
	H HandlerFunc
}

// ServeHTTP allows our Handler type to satisfy http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(h.Env, w, r)
	if err != nil {
		switch e := err.(type) {
		case StatusData:
			w.WriteHeader(e.Status())
			jsonData, _ := e.GetJsonData()
			w.Write(jsonData)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}
