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
	"sync"
	"time"

	"github.com/SanketJawali/naamon/src/db"
)

type HandlerFunc struct {
	Client       *http.Client
	ServerList   *ServerListT
	RateLimiters map[string]*RateLimiter
	Ctx          context.Context
	DB           *db.Queries
}

type ServerListT struct {
	List map[string]*db.GetApiMapByKeyRow
	mu   sync.RWMutex
}

type Policies struct {
	RateLimit *RateLimitPolicy `json:"rate_limit,omitempty"`
	// Auth      *AuthPolicy      `json:"auth,omitempty"`
	Timeout *TimeoutPolicy `json:"timeout,omitempty"`
}

type RateLimitPolicy struct {
	Capacity int `json:"capacity"`
	Rate     int `json:"rate"`
}

type TimeoutPolicy struct {
	DurationMs int `json:"duration_ms"`
}

type RateLimiter struct {
	Rate       float64
	Tokens     float64
	Capacity   int
	LastRefill time.Time
	mu         sync.Mutex
}

// Returns the Rate limiter instance if already exists, if not then returns a new Rate Limiter instance.
func (handler *HandlerFunc) GetRateLimiter(r *http.Request, rate float64, capacity int) *RateLimiter {
	// Rate limit with IP with stripped port
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	rateLimiter, exists := handler.RateLimiters[string(ip)]
	if !exists {
		log.Println("Returning a new rate limiter for IP: ", ip)
		newRl := &RateLimiter{
			Rate:       rate,
			Tokens:     float64(capacity),
			Capacity:   capacity,
			LastRefill: time.Now(),
		}
		handler.RateLimiters[string(ip)] = newRl
		return newRl
	}

	return rateLimiter
}

// Refills the tokens in token bucket for a client
func (rl *RateLimiter) Refill() {
	// To refill we calculate the time elapsed and set the tokens to the amount of tokens
	// that should have been added according to the rate
	// More efficient than having a thread adding new tokens time to time constantly
	now := time.Now()

	elapsed := now.Sub(rl.LastRefill).Seconds()
	rl.Tokens += elapsed * rl.Rate

	if rl.Tokens > float64(rl.Capacity) {
		rl.Tokens = float64(rl.Capacity)
	}

	rl.LastRefill = now
}

// Checks if the request being made can be accepted and isn't exceeding the rate limit
func (rl *RateLimiter) Accept() bool {
	// Refill the token bucket
	// If tokens available, accept and return true

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.Refill()

	if rl.Tokens >= 1 {
		rl.Tokens--
		return true
	}

	return false
}

// Get the API Entry and its data from the in memory cache
func (list *ServerListT) GetServerEntry(key string) (db.GetApiMapByKeyRow, bool) {
	list.mu.RLock()
	defer list.mu.RUnlock()

	entry, exists := list.List[key]
	if !exists {
		return db.GetApiMapByKeyRow{}, false
	}

	return *entry, exists
}

// Add an API Entry in the in memory cache
func (list *ServerListT) AddServerEntry(key string, value db.GetApiMapByKeyRow) bool {
	list.mu.Lock()
	defer list.mu.Unlock()

	// Check if exists
	_, exists := list.List[key]
	if exists {
		return false
	}

	// Add new server entry in server list
	list.List[key] = &value

	return true
}

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
	log.Println("URL Path: ", urlPath)
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
	apiEntry, exists := handler.ServerList.GetServerEntry(targetServer)
	if !exists {
		var err error
		dbEntry, err := handler.DB.GetApiMapByKey(handler.Ctx, targetId)
		if err != nil {
			log.Printf("Err: Error fetching target server for ID '%s': %v\n", targetId, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		targetServer = dbEntry.TargetUrl
		handler.ServerList.AddServerEntry(targetServer, dbEntry)
		apiEntry = dbEntry
	} else {
		targetServer = apiEntry.TargetUrl
	}

	if targetServer == "" {
		log.Printf("No target server found for ID '%s'\n", targetId)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 3. Properly construct the URL, including query parameters
	// Get the query parameters and append them to the target URL if they exist
	var targetUrl string

	if r.URL.RawQuery != "" {
		targetUrl = fmt.Sprintf("%v%s?%s", targetServer, targetRoute, r.URL.RawQuery)
	} else {
		targetUrl = fmt.Sprintf("%v%s", targetServer, targetRoute)
	}

	// Trauncate very long URLs, log the server we're routing to
	if len(targetUrl) > 80 {
		log.Printf("Routing request to: %s (truncated)\n", targetUrl[:80])
	} else {
		log.Println("Routing request to: ", targetUrl)
	}

	// Extract policies from the database entry and unmarshal them into the Policies struct
	var policies Policies
	log.Printf("Unmarshaling policies for ID '%s': %v\n", targetId, apiEntry.Policies.Valid)
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
	if (RateLimitPolicy{}) != *policies.RateLimit {

		rl := handler.GetRateLimiter(r, float64(policies.RateLimit.Rate), policies.RateLimit.Capacity)
		if !rl.Accept() {
			log.Printf("Request to API with Id %s canceled, Rate limit exceeded for IP %s\n", targetServer, r.RemoteAddr)
			http.Error(w, "Request to API canceled. Rate Limited", http.StatusTooManyRequests)
			return
		}
	}

	// 7. Make the request to the backend server
	resp, err := handler.Client.Do(proxyReq)
	if err != nil {
		// Check if the error occured due to a timeout
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("Err: Request to backend server at '%v' timed out: %v", targetServer, err)
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
