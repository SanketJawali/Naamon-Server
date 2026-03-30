package db

import (
	"context"
	"database/sql"
	"encoding/json"
)

func makePolicies(p map[string]interface{}) sql.NullString {
	if p == nil {
		return sql.NullString{String: "{}", Valid: true}
	}

	b, err := json.Marshal(p)
	if err != nil {
		return sql.NullString{}
	}

	return sql.NullString{
		String: string(b),
		Valid:  true,
	}
}

func (q *Queries) AddDummyData(ctx context.Context) error {

	data := []CreateApiMapParams{
		{
			UserID:    1,
			Key:       "abc123",
			TargetUrl: "http://localhost:5000",
			Policies: makePolicies(map[string]interface{}{
				"rate_limit": map[string]interface{}{
					"limit":  100,
					"window": "60s",
				},
			}),
		},
		{
			UserID:    1,
			Key:       "def456",
			TargetUrl: "http://localhost:5001",
			Policies: makePolicies(map[string]interface{}{
				"auth": map[string]interface{}{
					"enabled": true,
					"type":    "api_key",
				},
			}),
		},
		{
			UserID:    1,
			Key:       "ghi789",
			TargetUrl: "http://localhost:5002",
			Policies: makePolicies(map[string]interface{}{
				"rate_limit": map[string]interface{}{
					"limit":  50,
					"window": "30s",
				},
				"auth": map[string]interface{}{
					"enabled": true,
					"type":    "api_key",
				},
			}),
		},
		{
			UserID:    1,
			Key:       "jkl012",
			TargetUrl: "http://localhost:5003",
			Policies: makePolicies(map[string]interface{}{
				"timeout": map[string]interface{}{
					"duration_ms": 2000,
				},
			}),
		},
		{
			UserID:    2,
			Key:       "mno345",
			TargetUrl: "http://localhost:5004",
			Policies: makePolicies(map[string]interface{}{
				"rate_limit": map[string]interface{}{
					"limit":  10,
					"window": "10s",
				},
				"timeout": map[string]interface{}{
					"duration_ms": 1000,
				},
			}),
		},
		{
			UserID:    2,
			Key:       "pqr678",
			TargetUrl: "http://localhost:5005",
			Policies:  makePolicies(nil), // no policies
		},
	}

	for _, t := range data {
		if err := q.CreateApiMap(ctx, t); err != nil {
			return err
		}
	}

	return nil
}
