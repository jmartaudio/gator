-- name: GetFeeds :many
SELECT
    id,
    created_at,
    updated_at,
    name,
    url,
    user_id
FROM feeds WHERE user_id = $1
ORDER BY name ASC;
