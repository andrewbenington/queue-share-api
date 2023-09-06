-- name: AllRooms :many
SELECT *
FROM rooms;

-- name: FindRoomByCode :one
SELECT *
FROM rooms
WHERE (code = $1);

-- name: InsertRoom :one
INSERT INTO rooms (name)
VALUES ($1)
RETURNING *;