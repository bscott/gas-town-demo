package main

import (
	"log"
	"net/http"

	"gastowndemo/handlers"
)

func main() {
	api := handlers.NewAPI()
	ws := handlers.NewWSHandler()

	mux := http.NewServeMux()
	api.RegisterRoutes(mux)
	ws.RegisterRoutes(mux)

	log.Println("SlackLite server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
