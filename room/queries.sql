-- name: AllRooms :many
SELECT id, name, code, created
FROM rooms;

-- name: FindRoomByCode :one
SELECT id, name, code, created
FROM rooms
WHERE (code = $1);

-- name: GetRoomIDByCode :one
SELECT id
FROM rooms
WHERE (code = $1);

-- name: InsertRoom :one
INSERT INTO rooms (name, encrypted_access_token, access_token_expiry, encrypted_refresh_token)
VALUES ($1, $2, $3, $4)
RETURNING id, name, code, created;

-- name: GetRoomAuthByCode :one
SELECT encrypted_access_token, access_token_expiry, encrypted_refresh_token
FROM rooms
WHERE (code = $1);