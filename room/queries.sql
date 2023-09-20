-- name: FindRoomByCode :one
SELECT
    r.id,
    r.name,
    r.host_id,
    u.name as host_name,
    r.code,
    r.created
FROM rooms AS r
JOIN users AS u
    ON r.code = $1
    AND u.id = r.host_id;

-- name: GetRoomIDByCode :one
SELECT id
FROM rooms
WHERE (code = $1);

-- name: InsertRoom :one
INSERT INTO rooms (name, host_id)
VALUES ($1, $2)
RETURNING id, name, host_id, code, created;

-- name: GetSpotifyTokensByRoomCode :one
SELECT 
    st.encrypted_access_token,
    st.access_token_expiry,
    st.encrypted_refresh_token
FROM spotify_tokens AS st
JOIN rooms AS r
    ON r.code = $1
    AND u.host_id = r.host_id;

-- name: UpdateSpotifyTokensByRoomCode :exec
UPDATE spotify_tokens st
SET 
    encrypted_access_token = $2,
    access_token_expiry = $3,
    encrypted_refresh_token = $4
FROM rooms AS r
WHERE st.user_id = r.host_id
    AND r.code = $1;