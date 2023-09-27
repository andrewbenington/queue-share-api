-- name: UserUpdateSpotifyTokens :exec
INSERT INTO spotify_tokens(user_id, encrypted_access_token, access_token_expiry, encrypted_refresh_token)
    VALUES ($1, $2, $3, $4)
ON CONFLICT ON CONSTRAINT spotify_tokens_user_id_key
    DO UPDATE SET
        encrypted_access_token = $2, access_token_expiry = $3, encrypted_refresh_token = $4
    WHERE
        spotify_tokens.user_id = $1;

-- name: InsertUserWithPass :one
WITH new_user AS (
INSERT INTO users(username, display_name)
        VALUES ($1, $2)
    RETURNING
        id, username, display_name, created)
    INSERT INTO user_passwords(user_id, encrypted_password)
    SELECT
        id,
        crypt(@user_pass, gen_salt('bf'))
    FROM
        new_user
    RETURNING (
        SELECT
            created
        FROM
            new_user);

-- name: ValidateUserPass :one
SELECT
    (encrypted_password = crypt(@user_pass, encrypted_password))
FROM
    user_passwords AS up
    JOIN users u ON u.id = up.user_id
        AND UPPER(u.username) = UPPER(@username::text);

-- name: GetUserRoom :one
SELECT
    r.id,
    r.name,
    r.code,
    r.created,
    u.id AS user_id,
    u.username,
    u.display_name,
    u.spotify_image_url
FROM
    rooms r
    JOIN users u ON r.host_id = u.id
        AND upper(u.username) = upper(@username::text);

-- name: UserGetByUsername :one
SELECT
    id,
    username,
    display_name
FROM
    users u
WHERE
    username = $1;

-- name: UserGetByID :one
SELECT
    id,
    username,
    display_name,
    spotify_account,
    spotify_name,
    spotify_image_url
FROM
    users u
WHERE
    id = $1;

-- name: UserUpdateSpotifyInfo :exec
UPDATE
    users
SET
    spotify_account = $2,
    spotify_name = $3,
    spotify_image_url = $4
WHERE
    id = $1;

