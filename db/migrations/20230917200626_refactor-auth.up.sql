CREATE EXTENSION pgcrypto;

CREATE TABLE users(
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    username text NOT NULL,
    display_name text NOT NULL,
    spotify_account text,
    spotify_name text,
    spotify_image_url text,
    created timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX username_case_insensitive ON users(UPPER(username));

CREATE TABLE user_passwords(
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    encrypted_password text NOT NULL
);

INSERT INTO users(id, username, display_name)
    VALUES ('00000000-0000-0000-0000-000000000000', 'default', 'Unknown User');

ALTER TABLE rooms
    ALTER COLUMN created SET NOT NULL,
    ADD COLUMN host_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000' REFERENCES users(id) ON DELETE CASCADE,
    DROP COLUMN encrypted_access_token,
    DROP COLUMN access_token_expiry,
    DROP COLUMN encrypted_refresh_token;

CREATE TABLE room_passwords(
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    room_id uuid NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    encrypted_password bytea
);

CREATE TABLE spotify_tokens(
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id uuid UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    encrypted_access_token bytea NOT NULL,
    access_token_expiry timestamp with time zone NOT NULL,
    encrypted_refresh_token bytea NOT NULL
);

