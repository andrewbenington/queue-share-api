ALTER TABLE spotify_album_cache
    ADD COLUMN upc TEXT,
    ADD COLUMN spotify_track_ids JSONB,
    ADD COLUMN track_isrcs JSONB;

