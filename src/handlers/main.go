package handlers

import (
	"fmt"
	"net/http"
)

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	// 1. Validate the request
	fmt.Println("Received request:", r.Method, r.URL.Path)
	fmt.Println("Headers:", r.Header)
	fmt.Println("Body:", r.Body)

	// 2. Apply policies (e.g., rate limiting, authentication)
	// 3. Create a context for the request

	// 4. Fire the request
	// 5. After a response is generated, apply response policies (e.g., logging, modifying headers)
	// 6. Send the response back to the client
}

func IndexRouteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		w.Write([]byte("Welcome to Naamon!"))
	case http.MethodPost:
		w.Write([]byte("POST request received at /"))
	}
}
