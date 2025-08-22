-- name: GetFeeds :many
SELECT name, url FROM feeds WHERE user_id = $1
ORDER BY name ASC;
