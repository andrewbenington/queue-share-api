ALTER TABLE room_members
    DROP CONSTRAINT no_duplicate_room_members;

ALTER TABLE spotify_tokens
    DROP COLUMN permissions_version;

DROP TABLE spotify_permissions_versions;

ALTER TABLE rooms
    DROP COLUMN password_protected,
    DROP COLUMN updated,
    DROP COLUMN is_open;

ALTER TABLE room_queue_tracks
    DROP COLUMN played;

