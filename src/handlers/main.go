package handlers

import (
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
)

type HandlerFunc struct {
	Client *http.Client
}

func (handler HandlerFunc) RequestHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Set an temperory variable to hold the allocated server for this request
	allocatedServer := "http://localhost:5000"

	// 2. Properly construct the URL, including query parameters
	// The old url creation logic was removed. It was very inefficient
	// creating a new URL for each request is unnecessary
	// we can just concatenate the base server URL with the request path and query parameters.
	//
	// The old method used to use a base URL, with port number, and then add the URL path and query params to it.
	// Get the query parameters and append them to the target URL if they exist
	var targetUrl string

	if r.URL.RawQuery != "" {
		targetUrl = fmt.Sprintf("%s%s?%s", allocatedServer, r.URL.Path, r.URL.RawQuery)
	} else {
		targetUrl = fmt.Sprintf("%s%s", allocatedServer, r.URL.Path)
	}

	// Trauncate very long URLs, log the server we're routing to
	if len(targetUrl) > 80 {
		log.Printf("Routing request to: %s (truncated)\n", targetUrl[:80])
	} else {
		log.Println("Routing request to: ", targetUrl)
	}

	// 3. Simplify request creation
	// Using NewRequestWithContext is best practice so the request cancels if the client disconnects early
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetUrl, r.Body)
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. Copy original request headers to the proxy request
	// maps.Copy is a convenient way to copy all headers without needing to loop through them manually
	// manually copying headers with .Header.Add does more things behind the scenes, like checking for correct header formatting
	// which is unnecessary here since we're just copying them as-is,
	// so using maps.Copy is more efficient and less error-prone.
	maps.Copy(proxyReq.Header, r.Header)

	// 5. Actually execute the request using an HTTP client
	// NOTE: Each request consumes a token from the bucket of the allocated server.
	// If the bucket is empty, it will block until a token is available.
	// Stays blocked until a token is available.

	// <-tokenBucket[allocatedServer]
	// defer func() {
	// 	// Return the token to the bucket after the request is done
	// 	tokenBucket[allocatedServer] <- 1
	// }()

	resp, err := handler.Client.Do(proxyReq)
	if err != nil {
		log.Printf("Error reaching backend server at '%s': %v", allocatedServer, err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway) // No more log.Fatal
		return
	}
	defer resp.Body.Close()

	// 6. Copy backend response headers back to the client response
	// Refer to point 4 for why maps.Copy is used here as well
	maps.Copy(w.Header(), resp.Header)

	// 7. Write the exact status code returned by the backend
	w.WriteHeader(resp.StatusCode)

	// 8. Stream the body directly to avoid blowing up memory
	// Copying the response body directly prevents loading the entire response into memory
	// which is crucial for large responses
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error streaming response body: %v", err)
	}
}

func (handler HandlerFunc) HandleRequest(w http.ResponseWriter, r *http.Request) {
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

func (handler HandlerFunc) IndexRouteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		w.Write([]byte("Welcome to Naamon!\n"))
	case http.MethodPost:
		w.Write([]byte("POST request received at /\n"))
	}
}
