package main

import (
	"log"
	"net/http"

	"example.com/demo/db"
	"example.com/demo/repository"
	"example.com/demo/router"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	postgres, err := db.ConnectPostgres()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.ClosePostgres(postgres); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	repository.RoomRepo = repository.NewRoomRepository(postgres)
	db.ConnectRedis()
	defer db.CloseRedis()

	return http.ListenAndServe(":8080", router.MainRouter())
}
