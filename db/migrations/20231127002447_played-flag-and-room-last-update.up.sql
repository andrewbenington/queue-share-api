ALTER TABLE room_queue_tracks
    ADD COLUMN played boolean NOT NULL DEFAULT FALSE;

ALTER TABLE rooms
    ADD COLUMN password_protected boolean NOT NULL DEFAULT TRUE,
    ADD COLUMN updated TIMESTAMP WITH TIME ZONE DEFAULT now(),
    ADD COLUMN is_open boolean NOT NULL DEFAULT TRUE;

CREATE TABLE spotify_permissions_versions(
    id bigserial PRIMARY KEY,
    description text NOT NULL
);

INSERT INTO spotify_permissions_versions(description)
    VALUES ('Permission to read user subscription details and read/update playback state'),
('Permission to read user suggested tracks');

ALTER TABLE spotify_tokens
    ADD COLUMN permissions_version bigint NOT NULL REFERENCES spotify_permissions_versions(id) DEFAULT 1;

ALTER TABLE room_members
    ADD CONSTRAINT no_duplicate_room_members UNIQUE (user_id, room_id);

