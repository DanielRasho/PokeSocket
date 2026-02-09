
-- name: InsertUser :one
INSERT INTO users (username)
VALUES (@username)
RETURNING id;

-- name: DeleteUser :exec
DELETE FROM users 
WHERE id = @id;