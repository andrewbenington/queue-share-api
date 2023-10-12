CREATE TABLE room_guests(
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    room_id uuid NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    name text NOT NULL
);

CREATE TABLE room_queue_tracks(
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    track_id text NOT NULL,
    guest_id uuid NOT NULL REFERENCES room_guests(id) ON DELETE CASCADE,
    room_id uuid NOT NULL REFERENCES rooms(id) ON DELETE CASCADE
);

