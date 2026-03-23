-- name: GetAllUsers :many
SELECT * FROM users;

-- name: GetUserById :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: GetTargetUrlByApiKey :one
SELECT * FROM api_maps WHERE key = ?;

------------------------------------------------------------

-- name: CreateUser :exec
INSERT INTO users (username, email, password) VALUES (?, ?, ?);

-- name: CreateApiMap :exec
INSERT INTO api_maps (key, target_url) VALUES (?, ?);

-- name: UpdateApiMapByKey :exec
UPDATE api_maps SET target_url = ? WHERE key = ?;
