package handlers

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/SanketJawali/naamon/src/db"
)

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
func (list *ServerListT) AddServerEntry(key string, value *db.GetApiMapByKeyRow) int8 {
	list.mu.Lock()
	defer list.mu.Unlock()

	// Check if exists
	_, exists := list.List[key]
	if exists {
		return 0
	}

	// Add new server entry in server list
	list.List[key] = value

	// Check if added successfully
	if _, exists = list.List[key]; exists {
		return 1
	}

	return -1
}

// TODO: Complete this implementation
func (queue *FetchQueue) AddFetch(targetId string) error {
	if _, exists := queue.List[targetId]; !exists {
		queue.mu.Lock()
		defer queue.mu.Unlock()

		// Check again to confirm that the
		if _, exists = queue.List[targetId]; exists {
			return nil
		}

	}
	return nil
}
