-- name: ListArticles :many
SELECT *
FROM articles
ORDER BY published_at DESC
LIMIT @limit OFFSET @offset;

-- name: ListUnreadArticles :many
SELECT *
FROM articles
WHERE read = FALSE
ORDER BY published_at DESC
LIMIT @limit OFFSET @offset;

-- name: ListReadArticles :many
SELECT *
FROM articles
WHERE read = TRUE
ORDER BY published_at DESC
LIMIT @limit OFFSET @offset;

-- name: ListBookmarkedArticles :many
SELECT *
FROM articles
WHERE bookmarked = TRUE
ORDER BY published_at DESC
LIMIT @limit OFFSET @offset;

-- name: SearchArticles :many
SELECT *
FROM articles
WHERE title LIKE '%' || @query || '%'
   OR body LIKE '%' || @query || '%'
   OR details LIKE '%' || @query || '%'
   OR link LIKE '%' || @query || '%'
ORDER BY published_at DESC
LIMIT @limit OFFSET @offset;

-- name: Article :one
SELECT *
FROM articles
WHERE id = @id;

-- name: ArticleIDByGUID :one
SELECT id
FROM articles
WHERE guid = @guid;

-- name: NextUnreadArticle :one
SELECT *
FROM articles
WHERE read = FALSE
ORDER BY published_at DESC
LIMIT 1;

-- name: CreateArticle :one
INSERT OR IGNORE INTO articles (guid, title, body, published_at, link, feed, details, read, bookmarked)
VALUES (@guid, @title, @body, @published_at, @link, @feed, @details, @read, @bookmarked)
RETURNING id;

-- name: SetArticle :one
INSERT OR
REPLACE INTO articles (guid, title, body, published_at, link, feed, details, read, bookmarked)
VALUES (@guid, @title, @body, @published_at, @link, @feed, @details, @read, @bookmarked)
RETURNING id;

-- name: MarkReadArticle :exec
UPDATE articles
SET read = TRUE
WHERE id = @id;

-- name: MarkReadAllArticles :exec
UPDATE articles
SET read = TRUE
WHERE read = FALSE;

-- name: BookmarkArticle :exec
UPDATE articles
SET bookmarked = TRUE
WHERE id = @id;

-- name: UnbookmarkArticle :exec
UPDATE articles
SET bookmarked = FALSE
WHERE id = @id;

-- name: FindSession :one
SELECT *
FROM sessions
WHERE token = @token;

-- name: CommitSession :exec
INSERT OR
REPLACE INTO sessions (token, data, expiry)
VALUES (@token, @data, @expiry);

-- name: DeleteSession :exec
DELETE
FROM sessions
WHERE token = @token;