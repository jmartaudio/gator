-- name: GetFeedFollowsForUser :many
SELECT 
    ff.id,
    ff.created_at,
    ff.updated_at,
    ff.user_id,
    ff.feed_id,
    f.name AS feed_name,
    u.name AS user_name
FROM feed_follows AS ff
INNER JOIN feeds AS f ON ff.feed_id = f.id
INNER JOIN users AS u ON ff.user_id = u.id
WHERE ff.user_id = $1
ORDER BY f.name ASC;
