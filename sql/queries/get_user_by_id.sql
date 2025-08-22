-- name: GetUsernameById :one
SELECT name FROM users WHERE id = $1;
