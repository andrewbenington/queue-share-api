-- name: RoomGetByCode :one
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

-- name: RoomGetIDByCode :one
SELECT
    id
FROM
    rooms
WHERE (code = $1);

-- name: RoomInsertWithPassword :one
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

-- name: RoomValidatePassword :one
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

-- name: RoomUpdateSpotifyTokens :exec
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

-- name: RoomGuestInsert :one
INSERT INTO room_guests(room_id, name)
SELECT
    r.id,
    $1
FROM
    rooms AS r
WHERE
    r.code = @room_code::text
RETURNING
    id,
    name;

-- name: RoomGuestInsertWithID :one
INSERT INTO room_guests(id, room_id, name)
SELECT
    @guest_id::uuid,
    r.id,
    $1
FROM
    rooms AS r
WHERE
    r.code = @room_code::text
RETURNING
    id,
    name;

-- name: RoomGuestGetName :one
SELECT
    name
FROM
    room_guests
WHERE
    room_id = $1
    AND id = @guest_id::uuid;

-- name: RoomGetAllGuests :many
SELECT
    rg.name,
    rg.id,
    counts.queued_tracks
FROM
    room_guests AS rg
    JOIN rooms r ON r.code = $1
        AND rg.room_id = r.id
    JOIN (
        SELECT
            guest_id,
            COUNT(*) AS queued_tracks
        FROM
            room_queue_tracks
        GROUP BY
            guest_id) counts ON rg.id = counts.guest_id;

-- name: RoomSetGuestQueueTrack :exec
INSERT INTO room_queue_tracks(track_id, guest_id, room_id)
SELECT
    $1,
    @guest_id::uuid,
    r.id
FROM
    rooms AS r
WHERE
    r.code = @room_code::text;

-- name: RoomSetMemberQueueTrack :exec
INSERT INTO room_queue_tracks(track_id, user_id, room_id)
SELECT
    $1,
    @user_id::uuid,
    r.id
FROM
    rooms AS r
WHERE
    r.code = @room_code::text;

-- name: RoomGetQueueTracks :many
SELECT
    track_id,
    g.name AS guest_name,
    u.display_name AS member_name
FROM
    room_queue_tracks t
    JOIN rooms r ON r.code = $1
    LEFT JOIN room_guests g ON g.id = t.guest_id
    LEFT JOIN users u ON u.id = t.user_id;

-- name: RoomGetHostID :one
SELECT
    r.host_id
FROM
    rooms r
WHERE
    r.code = $1;

-- name: RoomDeleteByID :exec
DELETE FROM rooms r
WHERE r.code = $1;

-- name: RoomAddMember :exec
INSERT INTO room_members(user_id, room_id)
SELECT
    $1,
    r.id
FROM
    rooms AS r
WHERE
    r.code = @room_code::text;

-- name: RoomUserIsMember :one
SELECT
    is_moderator
FROM
    room_members
WHERE
    user_id = $1
    AND room_id = $2;

-- name: RoomGetAllMembers :many
SELECT
    u.id AS user_id,
    u.username,
    u.display_name,
    u.spotify_name,
    u.spotify_image_url,
    m.is_moderator,
    counts.queued_tracks
FROM
    room_members AS m
    JOIN users u ON m.user_id = u.id
        AND m.room_id = $1
    JOIN (
        SELECT
            user_id,
            COUNT(*) AS queued_tracks
        FROM
            room_queue_tracks
        GROUP BY
            user_id) counts ON u.id = counts.user_id;

-- name: RoomSetModerator :exec
UPDATE
    room_members
SET
    is_moderator = $3
WHERE
    room_id = $1
    AND user_id = $2;

