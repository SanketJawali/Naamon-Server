package handlers

import (
	"context"
	"net/http"
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
	FetchQueue   *FetchQueue
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
	Capacity int     `json:"capacity"`
	Rate     float64 `json:"rate"`
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

type FetchQueue struct {
	List map[string]*FetchQueueContent
	mu   sync.RWMutex
}

type FetchQueueContent struct {
	wg  *sync.WaitGroup
	err error
}
