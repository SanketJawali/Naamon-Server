package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/SanketJawali/naamon/src/db"
)

func setupTestDB(t *testing.T) (*db.Queries, context.Context) {
	t.Helper()

	ctx := context.Background()
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})

	schemaPath := filepath.Join("..", "db", "schema.sql")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schema file %q: %v", schemaPath, err)
	}

	if _, err := conn.ExecContext(ctx, string(schemaBytes)); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	queries := db.New(conn)
	if err := queries.CreateUser(ctx, db.CreateUserParams{
		Username: "tester",
		Email:    "tester@example.com",
		Password: "secret",
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	return queries, ctx
}

func newTestHandler(queries *db.Queries, ctx context.Context) *HandlerFunc {
	return &HandlerFunc{
		Client: &http.Client{},
		ServerList: &ServerListT{
			List: make(map[string]*db.GetApiMapByKeyRow),
		},
		RateLimiters: make(map[string]*RateLimiter),
		Ctx:          ctx,
		DB:           queries,
	}
}

func TestRequestHandler_ProxiesRequest(t *testing.T) {
	queries, ctx := setupTestDB(t)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "ok")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, "method=%s path=%s query=%s header=%s", r.Method, r.URL.Path, r.URL.RawQuery, r.Header.Get("X-Test"))
	}))
	t.Cleanup(backend.Close)

	if err := queries.CreateApiMap(ctx, db.CreateApiMapParams{
		UserID:    1,
		Key:       "abc123",
		TargetUrl: backend.URL,
		Policies:  sql.NullString{String: `{}`, Valid: true},
	}); err != nil {
		t.Fatalf("create api map: %v", err)
	}

	handler := newTestHandler(queries, ctx)
	req := httptest.NewRequest(http.MethodGet, "/abc123/health?x=1", nil)
	req.RemoteAddr = "127.0.0.1:45678"
	req.Header.Set("X-Test", "forward-me")

	rr := httptest.NewRecorder()
	handler.RequestHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	if rr.Header().Get("X-Backend") != "ok" {
		t.Fatalf("expected backend header to be copied")
	}

	body := rr.Body.String()
	if !strings.Contains(body, "path=/health") || !strings.Contains(body, "query=x=1") || !strings.Contains(body, "header=forward-me") {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestRequestHandler_ReturnsNotFoundForUnknownKey(t *testing.T) {
	queries, ctx := setupTestDB(t)
	handler := newTestHandler(queries, ctx)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	req.RemoteAddr = "127.0.0.1:45678"
	rr := httptest.NewRecorder()

	handler.RequestHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestRequestHandler_AppliesRateLimitPolicy(t *testing.T) {
	queries, ctx := setupTestDB(t)

	var backendCalls int32
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&backendCalls, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(backend.Close)

	if err := queries.CreateApiMap(ctx, db.CreateApiMapParams{
		UserID:    1,
		Key:       "limited",
		TargetUrl: backend.URL,
		Policies: sql.NullString{
			String: `{"rate_limit":{"capacity":1,"rate":0}}`,
			Valid:  true,
		},
	}); err != nil {
		t.Fatalf("create api map: %v", err)
	}

	handler := newTestHandler(queries, ctx)

	firstReq := httptest.NewRequest(http.MethodGet, "/limited", nil)
	firstReq.RemoteAddr = "127.0.0.1:9999"
	firstRes := httptest.NewRecorder()
	handler.RequestHandler(firstRes, firstReq)

	secondReq := httptest.NewRequest(http.MethodGet, "/limited", nil)
	secondReq.RemoteAddr = "127.0.0.1:9999"
	secondRes := httptest.NewRecorder()
	handler.RequestHandler(secondRes, secondReq)

	if firstRes.Code != http.StatusOK {
		t.Fatalf("expected first request status %d, got %d", http.StatusOK, firstRes.Code)
	}

	if secondRes.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request status %d, got %d", http.StatusTooManyRequests, secondRes.Code)
	}

	if atomic.LoadInt32(&backendCalls) != 1 {
		t.Fatalf("expected backend to be called once, got %d", backendCalls)
	}
}

func TestRequestHandler_TimeoutTest(t *testing.T) {
	queries, ctx := setupTestDB(t)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("delay") == "" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello World!"))
			return
		}

		delayDuration, err := strconv.Atoi(r.Header.Get("delay"))
		if err != nil {
			t.Fatal("Failed to Convert")
		}

		// Convert the delay duration (in ms) to ns. Multiplying by time.Millisecond converts the duration into ns
		time.Sleep(time.Duration(delayDuration * int(time.Millisecond)))

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello World! \nDelay of %v seconds", time.Duration(delayDuration).Seconds())
	}))
	t.Cleanup(func() { backend.Close() })

	targetId := "abc123"

	if err := queries.CreateApiMap(ctx, db.CreateApiMapParams{
		UserID:    1,
		Key:       targetId,
		TargetUrl: backend.URL,
		Policies: sql.NullString{
			String: `{"timeout": {"duration_ms": 500}}`,
			Valid:  true,
		},
	}); err != nil {
		t.Fatalf("Err: failed to add api key map in db: %v", err)
		return
	}

	handler := newTestHandler(queries, ctx)

	firstReq := httptest.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/%s", targetId), nil)
	firstReq.Header.Add("delay", "0")
	firstReq.RemoteAddr = "127.0.0.1:9876"
	firstRes := httptest.NewRecorder()
	handler.RequestHandler(firstRes, firstReq)

	secondReq := httptest.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/%s", targetId), nil)
	secondReq.Header.Add("delay", "700")
	secondReq.RemoteAddr = "127.0.0.1:8990"
	secondRes := httptest.NewRecorder()
	handler.RequestHandler(secondRes, secondReq)

	if firstRes.Code != http.StatusOK {
		t.Fatal("expected a successful request for first request")
	}

	if secondRes.Code != http.StatusGatewayTimeout {
		t.Fatal("expected a timeout status code for second request")
	}
}

func TestAddServerEntry_EnsureCachingSafety(t *testing.T) {
	cache := ServerListT{
		List: make(map[string]*db.GetApiMapByKeyRow),
	}
	var wg sync.WaitGroup

	// Writers
	for i := range 50 {
		wg.Add(1)

		wg.Go(func() {
			defer wg.Done()
			cache.AddServerEntry(fmt.Sprintf("key-%v", i), &db.GetApiMapByKeyRow{
				Key:       fmt.Sprintf("key-%v", i),
				TargetUrl: "http://localhost:8000/",
			})
		})
	}

	// Readers
	for i := range 50 {
		wg.Add(1)

		wg.Go(func() {
			defer wg.Done()
			cache.GetServerEntry(fmt.Sprintf("key-%v", i))
		})
	}

	wg.Wait()

	// Verify all writes were successful
	for i := range 50 {
		getFuncVal, exists := cache.GetServerEntry(fmt.Sprintf("key-%v", i))
		directVal := cache.List[fmt.Sprintf("key-%v", i)]

		if !exists || getFuncVal != *directVal || getFuncVal.Key != fmt.Sprintf("key-%v", i) {
			t.Fatal("Adding new Server entry in cache failed.\n")
		}

	}
}

func fakeDb_GetApiEntry(t *testing.T, key string, dbCalls *int32) *db.GetApiMapByKeyRow {
	t.Helper()
	atomic.AddInt32(dbCalls, 1)
	return &db.GetApiMapByKeyRow{
		Key:       key,
		TargetUrl: "http://localhost:8000/",
	}
}

func TestAddServerEntry_CountDbCalls(t *testing.T) {
	cache := ServerListT{
		List: make(map[string]*db.GetApiMapByKeyRow),
	}

	key := "abc123"
	var dbCalls int32 = 0

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)

		wg.Go(func() {
			defer wg.Done()

			val, exists := cache.GetServerEntry("abc123")

			if !exists {
				val = *fakeDb_GetApiEntry(t, key, &dbCalls)
				added := cache.AddServerEntry(key, &val)
				if added < 0 {
					t.Fatal("Failed to add new server entry to cache")
				}
			}
		})
	}

	wg.Wait()

	if dbCalls != 1 {
		t.Fatalf("expected only single db call, got %v", dbCalls)
	}
}
