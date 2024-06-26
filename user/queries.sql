-- name: UserUpdateSpotifyTokens :exec
WITH latest_perm_version AS (
  SELECT
    id
  FROM
    spotify_permissions_versions
  ORDER BY
    id DESC
  LIMIT 1)
  INSERT INTO spotify_tokens(
    user_id,
    encrypted_access_token,
    access_token_expiry,
    encrypted_refresh_token,
    permissions_version)
  SELECT
    $1,
    $2,
    $3,
    $4,
    id
  FROM
    latest_perm_version
  ON CONFLICT ON CONSTRAINT spotify_tokens_user_id_key
    DO UPDATE SET
      encrypted_access_token = $2,
      access_token_expiry = $3,
      encrypted_refresh_token = $4
    WHERE
      spotify_tokens.user_id = $1;

-- name: UserInsertWithPassword :one
WITH new_user AS (
  INSERT INTO users(
    username,
    display_name)
  VALUES (
    $1,
    $2)
RETURNING
  id,
  username,
  display_name,
  created)
INSERT INTO user_passwords(
  user_id,
  encrypted_password)
SELECT
  id,
  crypt(@user_pass, gen_salt('bf'))
FROM
  new_user
RETURNING (
  SELECT
    id
  FROM
    new_user);

-- name: UserUpdatePassword :exec
UPDATE
  users
SET
  encrypted_password = crypt(@user_pass, gen_salt('bf'))
WHERE
  id = @user_id;

-- name: UserValidatePassword :one
SELECT
  (encrypted_password = crypt(@user_pass, encrypted_password))
FROM
  user_passwords AS up
  JOIN users u ON u.id = up.user_id
    AND UPPER(u.username) = UPPER(@username::text);

-- name: UserGetHostedRooms :many
SELECT
  r.id,
  r.name,
  r.code,
  r.created,
  u.id AS host_id,
  u.username AS host_username,
  u.display_name AS host_display_name,
  u.spotify_image_url AS host_spotify_image_url
FROM
  rooms r
  JOIN users u ON r.host_id = u.id
    AND u.id = $1
    AND r.is_open = $2;

-- name: UserGetByUsername :one
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

-- name: UserDeleteSpotifyInfo :exec
UPDATE
  users
SET
  spotify_account = NULL,
  spotify_name = NULL,
  spotify_image_url = NULL
WHERE
  id = $1;

-- name: UserDeleteSpotifyToken :exec
DELETE FROM spotify_tokens
WHERE user_id = $1;

-- name: UserGetJoinedRooms :many
SELECT
  r.id,
  r.name,
  r.code,
  r.created,
  u.id AS host_id,
  u.username AS host_username,
  u.display_name AS host_display_name,
  u.spotify_image_url AS host_spotify_image_url
FROM
  rooms r
  JOIN room_members rm ON rm.user_id = $1
    AND r.id = rm.room_id
    AND r.is_open = $2
  JOIN users u ON r.host_id = u.id;

-- name: UserGetSpotifyTokens :one
SELECT
  st.encrypted_access_token,
  st.access_token_expiry,
  st.encrypted_refresh_token
FROM
  spotify_tokens AS st
WHERE
  st.user_id = $1;

-- name: UserHasSpotifyHistory :one
SELECT
  EXISTS (
    SELECT
      *
    FROM
      SPOTIFY_HISTORY
    WHERE
      USER_ID = @user_id
      AND from_history = TRUE);

-- name: UserGetAllWithSpotify :many
SELECT
  *
FROM
  users
WHERE
  spotify_account IS NOT NULL;

-- name: UserGetFriendSuggestions :many
SELECT
  id,
  username,
  display_name,
  spotify_image_url
FROM
  users u
WHERE
  NOT EXISTS (
    SELECT
      *
    FROM
      user_friend_requests fr
    WHERE
      fr.user_id = @user_id
      AND fr.friend_id = u.id)
  AND NOT EXISTS (
    SELECT
      *
    FROM
      user_friend_requests fr
    WHERE
      fr.friend_id = @user_id
      AND fr.user_id = u.id)
  AND NOT EXISTS (
    SELECT
      *
    FROM
      user_friends f
    WHERE
      f.user_id = @user_id
      AND f.friend_id = u.id)
  AND u.id != @user_id;

-- name: UserGetReceivedFriendRequests :many
SELECT
  id,
  username,
  display_name,
  spotify_image_url,
  request_timestamp
FROM
  user_friend_requests fr
  JOIN users u ON fr.user_id = u.id
    AND @user_id = fr.friend_id;

-- name: UserGetSentFriendRequests :many
SELECT
  id,
  username,
  display_name,
  spotify_image_url,
  request_timestamp
FROM
  user_friend_requests fr
  JOIN users u ON fr.friend_id = u.id
    AND @user_id = fr.user_id;

-- name: UserGetFriendRequestExists :one
SELECT
  EXISTS (
    SELECT
      *
    FROM
      user_friend_requests
    WHERE
      user_id = @user_id
      AND friend_id = @friend_id);

-- name: UserSendFriendRequest :exec
INSERT INTO user_friend_requests(
  user_id,
  friend_id)
VALUES (
  @user_id,
  @friend_id);

-- name: UserDeleteFriendRequest :exec
DELETE FROM user_friend_requests
WHERE user_id = @user_id
  AND friend_id = @friend_id;

-- name: UserInsertFriend :exec
INSERT INTO user_friends(
  user_id,
  friend_id)
VALUES (
  @user_id,
  @friend_id);

-- name: UserGetFriends :many
SELECT
  u.*
FROM
  user_friends f
  JOIN users u ON u.id = f.friend_id
    AND f.user_id = @user_id;

-- name: UserIsFriends :one
SELECT
  EXISTS (
    SELECT
      *
    FROM
      user_friends
    WHERE
      user_id = @user_id
      AND friend_id = @friend_id);

-- name: UsersToFetchHistory :many
SELECT
  u.*
FROM
  users u
  JOIN spotify_tokens st ON u.id = st.user_id
    AND st.permissions_version >= 3;

