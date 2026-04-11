package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"net"
	"net/http"
	"strings"
	"time"
)

const defaultProxyTimeout = 30 * time.Second

// Default handler which handles the requests made on the index route with an id
func (handler HandlerFunc) RequestHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extracting the Target ID, and route from the URL path
	// The URL path is expected to be in the format /{id}/... where {id} is the target ID
	if len(r.URL.Path) < 2 {
		http.Error(w, "No ID provided", http.StatusBadRequest)
		return
	}

	urlPath := r.URL.Path[1:]
	// Remove leading and trailing slashes to get the target ID
	urlSplit := strings.SplitN(urlPath, "/", 2)

	if len(urlSplit) < 1 {
		http.Error(w, "No ID Provided", http.StatusBadRequest)
	} else if len(urlSplit) < 2 {
		// Assuming that the Id is provided without the trailing `/` for index route
		// append the `/` for directing request to index route
		urlSplit = append(urlSplit, "/")
	}

	targetId := urlSplit[0]
	targetRoute := "/" + urlSplit[1]

	// Validate
	if targetId == "" || strings.Contains(targetId, "/") {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// 2. Get the target server from DB or cache
	var targetServer string

	// Check if the target server URL is already in the cache (ServerList)
	apiEntry, exists := handler.ServerList.GetServerEntry(targetId)
	if !exists {
		var err error
		dbEntry, err := handler.DB.GetApiMapByKey(handler.Ctx, targetId)
		if dbEntry.Key == "" || dbEntry.ID == 0 {
			log.Printf("Warning: No database entry found for ID '%v'\n", dbEntry)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if dbEntry.TargetUrl == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err != nil {
			log.Printf("Err: Error fetching target server for ID '%s': %v\n", targetId, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		targetServer = dbEntry.TargetUrl
		if status := handler.ServerList.AddServerEntry(targetId, &dbEntry); status < 0 {
			log.Printf("Err: Failed to add server entry for ID '%s' to cache\n", targetId)
		}
		apiEntry = dbEntry
	} else {
		targetServer = apiEntry.TargetUrl
	}

	// 3. Properly construct the URL, including query parameters
	// Get the query parameters and append them to the target URL if they exist
	var targetUrl string

	if r.URL.RawQuery != "" {
		targetUrl = fmt.Sprintf("%v%s?%s", targetServer, targetRoute, r.URL.RawQuery)
	} else {
		targetUrl = fmt.Sprintf("%v%s", targetServer, targetRoute)
	}

	// Extract policies from the database entry and unmarshal them into the Policies struct
	var policies Policies
	if !apiEntry.Policies.Valid {
		log.Printf("No valid policies found for ID '%s', using default policies\n", targetId)
	} else {
		err := json.Unmarshal([]byte(strings.TrimSpace(apiEntry.Policies.String)), &policies)
		if err != nil {
			log.Printf("Err: Error unmarshaling policies for ID '%s': %v\n", targetId, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// 4. Simplify request creation
	// Creating a new request context, with timeout policy added to it.
	timeout := defaultProxyTimeout
	if policies.Timeout != nil && policies.Timeout.DurationMs > 0 {
		timeout = time.Duration(policies.Timeout.DurationMs) * time.Millisecond
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	// Using NewRequestWithContext is best practice so the request cancels if the client disconnects early
	proxyReq, err := http.NewRequestWithContext(ctxWithTimeout, r.Method, targetUrl, r.Body)
	if err != nil {
		log.Printf("Err: Error creating proxy request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Copy original request headers to the proxy request
	// maps.Copy is a convenient way to copy all headers without needing to loop through them manually
	// manually copying headers with .Header.Add does more things behind the scenes, like checking for correct header formatting
	// which is unnecessary here since we're just copying them as-is,
	// so using maps.Copy is more efficient and less error-prone.
	maps.Copy(proxyReq.Header, r.Header)

	// 6. Apply policies here (authentication, rate limiting, etc.)
	if policies.RateLimit != nil && (RateLimitPolicy{}) != *policies.RateLimit {
		rl := handler.GetRateLimiter(r, float64(policies.RateLimit.Rate), policies.RateLimit.Capacity)
		if !rl.Accept() {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			log.Printf("Rate Limit: Request to API with Id %s canceled, Rate limit exceeded for IP %s\n", targetServer, ip)
			http.Error(w, "Request to API canceled. Rate Limited", http.StatusTooManyRequests)
			return
		}
	}

	// 7. Make the request to the backend server
	// Trauncate very long URLs, log the server we're routing to
	if len(targetUrl) > 80 {
		log.Printf("Routing request to: %s (truncated)\n", targetUrl[:80])
	} else {
		log.Println("Routing request to: ", targetUrl)
	}

	resp, err := handler.Client.Do(proxyReq)
	if err != nil {
		// Check if the error occured due to a timeout
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("Gateway Timeout: Request to backend server at '%v' timed out: %v", targetServer, err)
			http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
			return
		}

		log.Printf("Err: Error reaching backend server at '%v': %v", targetServer, err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway) // No more log.Fatal
		return
	}
	defer resp.Body.Close()

	// 8. Copy backend response headers back to the client response
	// Refer to point 5 for why maps.Copy is used here as well
	maps.Copy(w.Header(), resp.Header)

	// 9. Write the exact status code returned by the backend
	w.WriteHeader(resp.StatusCode)

	// 10. Stream the body directly to avoid blowing up memory
	// Copying the response body directly prevents loading the entire response into memory
	// which is crucial for large responses
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Err: Error streaming response body: %v", err)
	}
}
