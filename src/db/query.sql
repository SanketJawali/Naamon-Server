-- ******************* User Queries *******************

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: GetUserById :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: CreateUser :exec
INSERT INTO users (username, email, password)
VALUES (?, ?, ?);

------------------------------------------------------------

-- name: GetApiMapByKey :one
SELECT id, user_id, key, target_url, policies
FROM api_maps
WHERE key = ?;

-- name: CreateApiMap :exec
INSERT INTO api_maps (user_id, key, target_url, policies)
VALUES (?, ?, ?, ?);

-- name: UpdateApiMapTargetByKey :exec
UPDATE api_maps
SET target_url = ?, updated_at = CURRENT_TIMESTAMP
WHERE key = ?;

-- name: UpdateApiMapPoliciesByKey :exec
UPDATE api_maps
SET policies = ?, updated_at = CURRENT_TIMESTAMP
WHERE key = ?;

-- name: GetApiMapsByUser :many
SELECT id, user_id, key, target_url, policies
FROM api_maps
WHERE user_id = ?;

-- name: DeleteApiMapByKey :exec
DELETE FROM api_maps WHERE key = ?;
