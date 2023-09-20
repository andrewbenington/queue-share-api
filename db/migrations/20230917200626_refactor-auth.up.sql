CREATE EXTENSION pgcrypto;

CREATE TABLE users (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name TEXT NOT NULL,
    created TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_passwords (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    encrypted_password TEXT NOT NULL
);

INSERT INTO users (id, name)
VALUES ('00000000-0000-0000-0000-000000000000', 'Unknown User');

ALTER TABLE rooms
ALTER COLUMN created SET NOT NULL,
ADD COLUMN host_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000' REFERENCES users(id) ON DELETE CASCADE,
DROP COLUMN encrypted_access_token,
DROP COLUMN access_token_expiry,
DROP COLUMN encrypted_refresh_token;

CREATE TABLE room_passwords (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    encrypted_password BYTEA
);

CREATE TABLE spotify_tokens (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    encrypted_access_token BYTEA NOT NULL,
    access_token_expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    encrypted_refresh_token BYTEA NOT NULL
)