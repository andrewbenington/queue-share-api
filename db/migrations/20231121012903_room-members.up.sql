CREATE TABLE room_members(
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id uuid NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    is_moderator boolean NOT NULL DEFAULT FALSE
);

ALTER TABLE room_queue_tracks
    ADD COLUMN timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    ALTER COLUMN guest_id DROP NOT NULL;

