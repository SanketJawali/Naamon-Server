package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/SanketJawali/naamon/src/handlers"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	PORT := os.Getenv("PORT")
	log.Println("Starting server at port ", PORT)

	// Initializing the Client instance
	// Used to forward the requests from the clients to the servers
	handler := &handlers.HandlerFunc{
		Client:     &http.Client{},
		ServerList: make(map[string]string),
	}

	// Initialize HTTP server and routes
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler.RequestHandler)
	// mux.HandleFunc("/proxy/:id", handler.RequestHandler)

	http.ListenAndServe(fmt.Sprintf(":%v", PORT), mux)
}
