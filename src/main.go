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

	// Initialize HTTP server and routes
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.IndexRouteHandler)
	mux.HandleFunc("/proxy/:id", handlers.HandleRequest)

	http.ListenAndServe(fmt.Sprintf(":%v", PORT), mux)
}
