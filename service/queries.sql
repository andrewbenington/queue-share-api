-- name: TrackCacheGetByID :many
SELECT
    *
FROM
    SPOTIFY_TRACK_CACHE
WHERE
    id = ANY (@track_ids::text[]);

-- name: TrackCacheInsertBulk :exec
INSERT INTO SPOTIFY_TRACK_CACHE(
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
    isrc)
VALUES (
    unnest(
        @id ::text[]),
    unnest(
        @uri ::text[]),
    unnest(
        @name ::text[]),
    unnest(
        @album_id ::text[]),
    unnest(
        @album_uri ::text[]),
    unnest(
        @album_name ::text[]),
    unnest(
        @artist_id ::text[]),
    unnest(
        @artist_uri ::text[]),
    unnest(
        @artist_name ::text[]),
    unnest(
        @image_url ::text[]),
    unnest(
        @other_artists ::jsonb[]),
    unnest(
        @duration_ms ::int[]),
    unnest(
        @popularity ::int[]),
    unnest(
        @explicit ::bool[]),
    unnest(
        @preview_url ::text[]),
    unnest(
        @disc_number ::int[]),
    unnest(
        @track_number ::int[]),
    unnest(
        @type ::text[]),
    unnest(
        @external_ids ::jsonb[]),
    unnest(
        @isrc ::text[]))
ON CONFLICT
    DO NOTHING;

-- name: AlbumCacheGetByID :many
SELECT
    *
FROM
    SPOTIFY_ALBUM_CACHE
WHERE
    id = ANY (@album_ids::text[]);

-- name: AlbumCacheInsertBulk :exec
INSERT INTO SPOTIFY_ALBUM_CACHE(
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
    popularity)
VALUES (
    unnest(
        @id ::text[]),
    unnest(
        @uri ::text[]),
    unnest(
        @name ::text[]),
    unnest(
        @artist_id ::text[]),
    unnest(
        @artist_uri ::text[]),
    unnest(
        @artist_name ::text[]),
    unnest(
        @album_group ::text[]),
    unnest(
        @album_type ::text[]),
    unnest(
        @image_url ::text[]),
    unnest(
        @release_date ::date[]),
    unnest(
        @release_date_precision ::text[]),
    unnest(
        @genres ::jsonb[]),
    unnest(
        @popularity ::int[]))
ON CONFLICT
    DO NOTHING;

-- name: ArtistCacheGetPresentIDs :many
SELECT
    id
FROM
    SPOTIFY_ARTIST_CACHE
WHERE
    id = ANY (@artist_ids::text[]);

-- name: ArtistCacheGetByID :many
SELECT
    *
FROM
    SPOTIFY_ARTIST_CACHE
WHERE
    id = ANY (@artist_ids::text[]);

-- name: ArtistCacheInsertBulk :copyfrom
INSERT INTO SPOTIFY_ARTIST_CACHE(
    id,
    uri,
    name,
    image_url,
    genres,
    popularity,
    follower_count)
VALUES (
    @id,
    @uri,
    @name,
    @image_url,
    @genres,
    @popularity,
    @follower_count);

-- name: ArtistCacheInsertOne :exec
INSERT INTO SPOTIFY_ARTIST_CACHE(
    id,
    uri,
    name,
    image_url,
    genres,
    popularity,
    follower_count)
VALUES (
    @id,
    @uri,
    @name,
    @image_url,
    @genres,
    @popularity,
    @follower_count);

-- name: ArtistCacheUpdateOne :exec
UPDATE
    SPOTIFY_ARTIST_CACHE
SET
    name = @name,
    image_url = @image_url,
    genres = @genres,
    popularity = @popularity,
    follower_count = @follower_count
WHERE
    id = @id;

-- name: AlbumCacheInsertOne :exec
INSERT INTO SPOTIFY_ALBUM_CACHE(
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
    popularity,
    upc,
    spotify_track_ids,
    track_isrcs)
VALUES (
    @id,
    @uri,
    @name,
    @artist_id,
    @artist_uri,
    @artist_name,
    @album_group,
    @album_type,
    @image_url,
    @release_date,
    @release_date_precision,
    @genres,
    @popularity,
    @upc,
    @spotify_track_ids,
    @track_isrcs);

-- name: AlbumCacheUpdateOne :exec
UPDATE
    SPOTIFY_ALBUM_CACHE
SET
    name = @name,
    image_url = @image_url,
    genres = @genres,
    popularity = @popularity,
    upc = @upc,
    spotify_track_ids = @spotify_track_ids,
    track_isrcs = @track_isrcs
WHERE
    id = @id;

-- name: TracksGetPrimaryURIs :many
WITH top_isrcs AS (
    SELECT
        tc.isrc,
(json_agg(DISTINCT tc.uri)) AS original_uris
    FROM
        spotify_track_cache tc
    WHERE
        tc.uri = ANY (@uris::text[])
    GROUP BY
        tc.isrc
),
pref_albums AS (
    SELECT DISTINCT ON (top_isrcs.isrc)
        top_isrcs.*,
        tc.uri AS primary_uri
    FROM
        top_isrcs
        JOIN spotify_track_cache tc ON tc.isrc = top_isrcs.isrc
        JOIN spotify_album_cache ac ON ac.id = tc.album_id
    ORDER BY
        top_isrcs.isrc,
        CASE WHEN ac.album_type = 'album' THEN
            1
        WHEN ac.album_type = 'single' THEN
            2
        WHEN ac.album_type = 'compilation' THEN
            3
        ELSE
            4
        END,
        release_date DESC
)
SELECT
    *
FROM
    pref_albums;

