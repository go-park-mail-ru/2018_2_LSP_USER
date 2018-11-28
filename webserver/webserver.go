package webserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2018_2_LSP_USER/webserver/handlers"
	"github.com/go-park-mail-ru/2018_2_LSP_USER/webserver/routes"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	zap "go.uber.org/zap"
)

// Run Run webserver on specified port (passed as string the
// way regular http.ListenAndServe works)
func Run(addr string) {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("Can't create logger", err)
		return
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	grcpUser, err := grpc.Dial(
		"user-grpc:8080",
		grpc.WithInsecure(),
	)
	if err != nil {
		sugar.Errorw("Can't connect to user grpc",
			"error", err,
		)
		return
	}
	defer grcpUser.Close()

	grcpAuth, err := grpc.Dial(
		"auth-grpc:8080",
		grpc.WithInsecure(),
	)

	if err != nil {
		sugar.Errorw("Can't connect to user grpc",
			"error", err,
		)
		return
	}
	defer grcpAuth.Close()

	env := &handlers.Env{
		Logger:   sugar,
		GRCPUser: grcpUser,
		GRCPAuth: grcpAuth,
	}

	http.Handle("/metrics", promhttp.Handler())

	handlersMap := routes.Get()
	for URL, h := range handlersMap {
		http.Handle(URL, handlers.Handler{env, h})
	}

	log.Fatal(http.ListenAndServe(addr, nil))
}
