package db

import (
	"context"
	"database/sql"
)

func (q *Queries) AddDummyData(ctx context.Context) error {
	data := []CreateApiMapParams{
		{
			Key:       "abc123",
			TargetUrl: sql.NullString{String: "http://localhost:5000", Valid: true},
		},
		{
			Key:       "def456",
			TargetUrl: sql.NullString{String: "http://localhost:5001", Valid: true},
		},
		{
			Key:       "ghi789",
			TargetUrl: sql.NullString{String: "http://localhost:5002", Valid: true},
		},
	}

	for _, t := range data {
		err := q.CreateApiMap(ctx, t)

		if err != nil {
			return err
		}
	}

	return nil
}
