package main

import (
	"strings"
	"testing"
)

func TestSchemaEmbedded(t *testing.T) {
	if strings.TrimSpace(schema) == "" {
		t.Fatal("embedded schema should not be empty")
	}

	if !strings.Contains(schema, "CREATE TABLE users") {
		t.Fatal("embedded schema is missing users table")
	}
}
