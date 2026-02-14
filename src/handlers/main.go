package handlers

import (
	"fmt"
	"net/http"
)

func IndexRouteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index route handler.\n")
}

func RegisterRouteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Register route handler.\n")
}

func LoginRouteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Login route handler.\n")
}

func SyncUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Sync User Handler.\n")
}
