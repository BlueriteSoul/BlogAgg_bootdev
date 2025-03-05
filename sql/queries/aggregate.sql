-- name: MarkFeedFetched :exec
UPDATE feeds
SET
    last_fetched_at = NOW(),
    updated_at = NOW()
WHERE
    id = $1;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at NULLS FIRST
LIMIT 1;

-- name: CreatePost :exec
INSERT INTO posts (
    id,
    created_at,
    updated_at,
    title,
    url,
    description,
    published_at,
    feed_id
) VALUES (
    $1,  -- id (UUID)
    $2,  -- created_at (TIMESTAMP)
    $3,  -- updated_at (TIMESTAMP)
    $4,  -- title (TEXT)
    $5,  -- url (TEXT)
    $6,  -- description (TEXT)
    $7,  -- published_at (TIMESTAMP)
    $8   -- feed_id (UUID)
);

-- name: GetPostsForUser :many
SELECT
    posts.*
FROM
    posts
JOIN
    feeds ON posts.feed_id = feeds.id
JOIN
    feed_follows ON feeds.id = feed_follows.feed_id
JOIN
    users ON feed_follows.user_id = users.id
WHERE
    users.name = $1
ORDER BY
    posts.published_at DESC
LIMIT
    $2;