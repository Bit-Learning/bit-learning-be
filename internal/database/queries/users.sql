-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name)
VALUES ($1, $2, $3)
RETURNING id, email, password_hash, display_name, bio, avatar_url, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, display_name, bio, avatar_url, created_at, updated_at
FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, display_name, bio, avatar_url, created_at, updated_at
FROM users WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET display_name = $2, bio = $3, avatar_url = $4, updated_at = now()
WHERE id = $1
RETURNING id, email, password_hash, display_name, bio, avatar_url, created_at, updated_at;
