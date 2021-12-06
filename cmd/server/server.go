package main

import (
	"flag"
	"github.com/antonjah/poopchat/internal/poopchat"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var address = flag.String("address", "0.0.0.0:8080", "listen bind address")

func main() {
	flag.Parse()

	server := poopchat.NewServer()
	go server.Run()

	router := mux.NewRouter()

	router.Use(poopchat.SessionIDMiddleware)

	// define handlers
	router.HandleFunc("/", poopchat.ChatHandler)
	router.HandleFunc("/socket", func(w http.ResponseWriter, r *http.Request) {
		poopchat.Serve(server, w, r)
	})

	// listen and server
	log.Infof("Poopchat running on %s", *address)
	if err := http.ListenAndServe(*address, router); err != nil {
		log.WithError(err).Error("Listen and serve failed")
	}
}
