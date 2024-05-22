-- name: TrackCacheGetByID :many
SELECT
    *
FROM
    SPOTIFY_TRACK_CACHE
WHERE
    id = ANY(@track_ids :: text []);

-- name: TrackCacheInsertBulk :exec
INSERT INTO
    SPOTIFY_TRACK_CACHE(
        id,
        uri,
        name,
        album_id,
        album_uri,
        album_name,
        artist_id,
        artist_uri,
        artist_name,
        image_url,
        other_artists,
        duration_ms,
        popularity,
        explicit,
        preview_url,
        disc_number,
        track_number,
        type,
        external_ids,
        isrc
    )
VALUES
    (
        unnest(@id :: text []),
        unnest(@uri :: TEXT []),
        unnest(@name :: TEXT []),
        unnest(@album_id :: TEXT []),
        unnest(@album_uri :: TEXT []),
        unnest(@album_name :: TEXT []),
        unnest(@artist_id :: TEXT []),
        unnest(@artist_uri :: TEXT []),
        unnest(@artist_name :: TEXT []),
        unnest(@image_url :: TEXT []),
        unnest(@other_artists :: JSONB []),
        unnest(@duration_ms :: int []),
        unnest(@popularity :: int []),
        unnest(@explicit :: bool []),
        unnest(@preview_url :: TEXT []),
        unnest(@disc_number :: int []),
        unnest(@track_number :: int []),
        unnest(@type :: TEXT []),
        unnest(@external_ids :: jsonb []),
        unnest(@isrc :: text [])
    ) ON CONFLICT DO NOTHING;

-- name: AlbumCacheInsertBulk :exec
INSERT INTO
    SPOTIFY_ALBUM_CACHE(
        id,
        uri,
        name,
        artist_id,
        artist_uri,
        artist_name,
        album_group,
        album_type,
        image_url,
        release_date,
        release_date_precision,
        genres,
        popularity
    )
VALUES
    (
        unnest(@id :: text []),
        unnest(@uri :: TEXT []),
        unnest(@name :: TEXT []),
        unnest(@artist_id :: TEXT []),
        unnest(@artist_uri :: TEXT []),
        unnest(@artist_name :: TEXT []),
        unnest(@album_group :: TEXT []),
        unnest(@album_type :: TEXT []),
        unnest(@image_url :: TEXT []),
        unnest(@release_date :: date []),
        unnest(@release_date_precision :: TEXT []),
        unnest(@genres :: jsonb []),
        unnest(@popularity :: int [])
    ) ON CONFLICT DO NOTHING;

-- name: ArtistCacheInsertBulk :exec
INSERT INTO
    SPOTIFY_ARTIST_CACHE(
        id,
        uri,
        name,
        image_url,
        genres,
        popularity,
        follower_count
    )
VALUES
    (
        unnest(@id :: text []),
        unnest(@uri :: TEXT []),
        unnest(@name :: TEXT []),
        unnest(@image_url :: TEXT []),
        unnest(@genres :: jsonb []),
        unnest(@popularity :: int []),
        unnest(@follower_count :: int [])
    ) ON CONFLICT DO NOTHING;