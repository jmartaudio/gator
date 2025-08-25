-- name: GetPostForUser :many
SELECT 
    id,
    created_at,
    updated_at,
    title,
    url,
    description,
    published_at,
    feed_id
FROM posts
WHERE feed_id = $1
ORDER BY updated_at ASC
LIMIT $2;
