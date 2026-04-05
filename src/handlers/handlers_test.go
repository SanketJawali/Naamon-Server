package handlers_test

import (
	"net/http/httptest"
	"testing"
)

func TestRequestHandler(t *testing.T) {
	server := httptest.NewServer(RequestHandler)
}
