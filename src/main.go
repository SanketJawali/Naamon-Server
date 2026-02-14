package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/tursodatabase/go-libsql"

	dbpkg "github.com/SanketJawali/naamon/src/db"
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

	// Database initialization, in embeded replica format
	dbName := "local.db"
	primaryUrl := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_DATABASE_AUTH_TOKEN")

	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		fmt.Println("Error creating temporary directory:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, dbName)

	connector, err := libsql.NewEmbeddedReplicaConnector(dbPath, primaryUrl,
		libsql.WithAuthToken(authToken),
	)
	if err != nil {
		fmt.Println("Error creating connector:", err)
		os.Exit(1)
	}
	defer connector.Close()

	db := sql.OpenDB(connector)
	defer db.Close()

	// Verify database tables
	err = dbpkg.CreateTables(db)
	if err != nil {
		panic("DB Table creation error.")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.IndexRouteHandler)
	mux.HandleFunc("/api/auth/register", handlers.RegisterRouteHandler)
	mux.HandleFunc("/api/auth/login", handlers.LoginRouteHandler)
	mux.HandleFunc("/api/sync-user", handlers.SyncUserHandler)

	http.ListenAndServe(fmt.Sprintf(":%v", PORT), mux)
}
