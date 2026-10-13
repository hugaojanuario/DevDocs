-- name: CreateSaved :one
INSERT INTO saved_queries (id, question, title, technology, response)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetSaved :one
SELECT * FROM saved_queries WHERE id = $1;

-- name: ListSaved :many
SELECT id, question, title, technology, created_at, updated_at FROM saved_queries
WHERE (sqlc.arg(filter)::text = '' OR strpos(lower(title || ' ' || question), lower(sqlc.arg(filter))) > 0)
AND (sqlc.arg(tech)::text = '' OR technology = sqlc.arg(tech))
ORDER BY created_at DESC, id DESC LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountSaved :one
SELECT count(*) FROM saved_queries
WHERE (sqlc.arg(filter)::text = '' OR strpos(lower(title || ' ' || question), lower(sqlc.arg(filter))) > 0)
AND (sqlc.arg(tech)::text = '' OR technology = sqlc.arg(tech));

-- name: RenameSaved :one
UPDATE saved_queries SET title = $2, updated_at = NOW() WHERE id = $1 RETURNING *;

-- name: DeleteSaved :execrows
DELETE FROM saved_queries WHERE id = $1;
