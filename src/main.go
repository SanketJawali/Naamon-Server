package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"

	"github.com/SanketJawali/naamon/src/db"
	"github.com/SanketJawali/naamon/src/handlers"
)

// Store schama in a variable during compile time.
//
//go:embed db/schema.sql
var schema string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	PORT := os.Getenv("PORT")
	log.Println("Starting server at port ", PORT)

	ctx := context.Background()

	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// create tables
	if _, err := conn.ExecContext(ctx, schema); err != nil {
		log.Fatal(err)
	}

	// pass *sql.DB into sqlc
	queries := db.New(conn)
	log.Println("Database initialized successfully")

	// Add some dummy data to the database
	log.Println("Adding dummy data to the database")
	if err := queries.AddDummyData(ctx); err != nil {
		log.Fatal(err)
	}

	// Initializing the Client instance
	// Used to forward the requests from the clients to the servers
	serverList := &handlers.ServerListT{
		List: make(map[string]*db.GetApiMapByKeyRow),
	}

	handler := &handlers.HandlerFunc{
		Client:       &http.Client{},
		ServerList:   serverList,
		RateLimiters: make(map[string]*handlers.RateLimiter),
		Ctx:          ctx,
		DB:           queries,
		FetchQueue:   make(map[string]*handlers.FetchQueue),
	}

	// Initialize HTTP server and routes
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler.RequestHandler)

	http.ListenAndServe(fmt.Sprintf(":%v", PORT), mux)
}
