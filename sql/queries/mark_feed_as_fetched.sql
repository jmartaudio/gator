-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1 AND updated_at = $2
WHERE id = $3;
