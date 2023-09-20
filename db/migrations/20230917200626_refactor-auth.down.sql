DROP TABLE spotify_tokens;

DROP TABLE room_passwords;

ALTER TABLE rooms
ALTER COLUMN created DROP NOT NULL,
DROP COLUMN host_id,
ADD COLUMN encrypted_access_token TEXT NOT NULL DEFAULT '',
ADD COLUMN access_token_expiry TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
ADD COLUMN encrypted_refresh_token TEXT NOT NULL DEFAULT '';

DROP TABLE user_passwords;

DROP TABLE users;

DROP EXTENSION pgcrypto;