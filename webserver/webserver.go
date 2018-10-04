package webserver

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-park-mail-ru/2018_2_LSP_USER/webserver/handlers"
	"github.com/go-park-mail-ru/2018_2_LSP_USER/webserver/routes"
)

// Run Run webserver on specified port (passed as string the
// way regular http.ListenAndServe works)
func Run(addr string, db *sql.DB) {
	env := &handlers.Env{
		DB:       db,
		Host:     os.Getenv("DB_HOST"),
		Database: os.Getenv("DB_DB"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}

	handlersMap := routes.Get()
	for URL, h := range handlersMap {
		http.Handle(URL, handlers.Handler{env, h})
	}
	log.Fatal(http.ListenAndServe(addr, nil))
}
