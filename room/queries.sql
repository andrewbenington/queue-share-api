-- name: FindRoomByCode :one
SELECT
    r.id,
    r.name,
    r.host_id,
    u.username AS host_username,
    u.display_name AS host_display,
    u.spotify_name AS host_spotify_name,
    u.spotify_image_url AS host_image,
    r.code,
    r.created
FROM
    rooms AS r
    JOIN users AS u ON r.code = $1
        AND u.id = r.host_id;

-- name: GetRoomIDByCode :one
SELECT
    id
FROM
    rooms
WHERE (code = $1);

-- name: InsertRoomWithPass :one
WITH new_room AS (
INSERT INTO rooms(name, host_id)
        VALUES ($1, $2)
    RETURNING
        id, name, host_id, code, created
), new_pass AS (
INSERT INTO room_passwords(room_id, encrypted_password)
    SELECT
        id,
        crypt(@room_pass, gen_salt('bf'))
    FROM
        new_room
)
SELECT
    *
FROM
    new_room;

-- name: ValidateRoomPass :one
SELECT
    (encrypted_password = crypt(@room_pass::text, encrypted_password::text))
FROM
    room_passwords AS rp
    JOIN rooms r ON r.id = rp.room_id
        AND r.code = $1;

-- name: GetSpotifyTokensByRoomCode :one
SELECT
    st.encrypted_access_token,
    st.access_token_expiry,
    st.encrypted_refresh_token
FROM
    spotify_tokens AS st
    JOIN rooms AS r ON r.code = $1
        AND st.user_id = r.host_id;

-- name: UpdateSpotifyTokensByRoomCode :exec
UPDATE
    spotify_tokens st
SET
    encrypted_access_token = $2,
    access_token_expiry = $3,
    encrypted_refresh_token = $4
FROM
    rooms AS r
WHERE
    st.user_id = r.host_id
    AND r.code = $1;

