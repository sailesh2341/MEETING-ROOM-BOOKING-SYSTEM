package main

import (
	"log"
	"net/http"

	"github.com/sailesh-kona/meeting-room-booking-system/config"
	"github.com/sailesh-kona/meeting-room-booking-system/db"
	"github.com/sailesh-kona/meeting-room-booking-system/server"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(database)

	log.Println("server listening on", cfg.Address)
	err = http.ListenAndServe(cfg.Address, server.New(database, cfg))
	if err != nil {
		log.Fatal(err)
	}
}
