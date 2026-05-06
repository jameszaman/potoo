-- name: CreateEnvironment :one
INSERT INTO environments (id, project_id, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetProductionEnvironment :one
SELECT * FROM environments
WHERE project_id = $1 AND name = 'production'
LIMIT 1;
