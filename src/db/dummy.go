package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
)

func makePolicies(p map[string]interface{}) sql.NullString {
	if p == nil {
		return sql.NullString{String: "{}", Valid: true}
	}

	b, err := json.Marshal(p)
	if err != nil {
		log.Printf("Error marshaling policies: %v", err)
		return sql.NullString{}
	}

	return sql.NullString{
		String: string(b),
		Valid:  bool(true),
	}
}

func (q *Queries) AddDummyData(ctx context.Context) error {

	data := []CreateApiMapParams{
		{
			UserID:    1,
			Key:       "abc123",
			TargetUrl: "http://localhost:8000",
			Policies: makePolicies(map[string]interface{}{
				"rate_limit": map[string]interface{}{
					"capacity": 5,
					"rate":     1.0,
				},
			}),
		},
		{
			UserID:    1,
			Key:       "def456",
			TargetUrl: "http://localhost:8000",
			Policies:  makePolicies(map[string]interface{}{}), // no policies
		},
		{
			UserID:    1,
			Key:       "ghi789",
			TargetUrl: "http://localhost:8000",
			Policies: makePolicies(map[string]interface{}{
				"rate_limit": map[string]interface{}{
					"capacity": 50,
					"rate":     1.67,
				},
			}),
		},
		{
			UserID:    1,
			Key:       "jkl012",
			TargetUrl: "http://localhost:8000",
			Policies: makePolicies(map[string]interface{}{
				"timeout": map[string]interface{}{
					"duration_ms": 2000,
				},
			}),
		},
		{
			UserID:    2,
			Key:       "mno345",
			TargetUrl: "http://localhost:8000",
			Policies: makePolicies(map[string]interface{}{
				"rate_limit": map[string]interface{}{
					"capacity": 10,
					"rate":     1.0,
				},
				"timeout": map[string]interface{}{
					"duration_ms": 1000,
				},
			}),
		},
	}

	for _, t := range data {
		err := q.CreateApiMap(ctx, t)
		if err != nil {
			return err
		}
		log.Printf("Added dummy data for key '%s' | Target: %s | Policies: %s\n", t.Key, t.TargetUrl, t.Policies.String)
	}

	return nil
}
