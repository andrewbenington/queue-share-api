DROP TABLE room_members;

ALTER TABLE room_queue_tracks
    DROP COLUMN timestamp,
    DROP COLUMN user_id,
    ALTER COLUMN guest_id SET NOT NULL;

