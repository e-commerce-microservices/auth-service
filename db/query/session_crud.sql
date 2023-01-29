-- name: CreateSession :exec
INSERT INTO session (
    "id","user_id", "refresh_token", "user_agent", "client_ip", "expires_at"
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: GetSession :one
SELECT * FROM session
WHERE refresh_token = $1 LIMIT 1;