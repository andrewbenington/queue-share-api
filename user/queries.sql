-- name: UpdateUserSpotifyTokens :exec
UPDATE spotify_tokens SET encrypted_access_token = $2, access_token_expiry = $3, encrypted_refresh_token = $4
WHERE user_id = $1;

-- name: InsertUserWithPass :one
WITH new_user AS (
    INSERT INTO users(name)
    VALUES ($1)
    RETURNING id, name, created
)
INSERT INTO user_passwords (user_id, encrypted_password)
SELECT id, crypt(@user_pass, gen_salt('bf')) FROM new_user
RETURNING (SELECT id FROM new_user);

-- name: ValidateUserPass :one
SELECT (encrypted_password = crypt(@user_pass, encrypted_password))
FROM user_passwords WHERE user_id = $1;