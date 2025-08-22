-- name: GetFeedsByUrl :one
SELECT id, created_at, updated_at, name, url, user_id FROM feeds WHERE url = $1;
