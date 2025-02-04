ALTER TABLE spotify_album_cache
    DROP COLUMN IF EXISTS spotify_track_ids,
    DROP COLUMN IF EXISTS track_isrcs;

