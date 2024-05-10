-- name: HistoryInsertOne :exec
INSERT INTO
    SPOTIFY_HISTORY(
        user_id,
        timestamp,
        platform,
        ms_played,
        conn_country,
        ip_addr,
        user_agent,
        track_name,
        artist_name,
        album_name,
        spotify_track_uri,
        spotify_artist_uri,
        spotify_album_uri,
        reason_start,
        reason_end,
        shuffle,
        skipped,
        offline,
        offline_timestamp,
        incognito_mode,
        from_history
    )
VALUES
    (
        @user_id,
        @timestamp,
        @platform,
        @ms_played,
        @conn_country,
        @ip_addr,
        @user_agent,
        @track_name,
        @artist_name,
        @album_name,
        @spotify_track_uri,
        @spotify_artist_uri,
        @spotify_album_uri,
        @reason_start,
        @reason_end,
        @shuffle,
        @skipped,
        @offline,
        @offline_timestamp,
        @incognito_mode,
        @from_history
    );

-- name: HistoryInsertBulk :exec
INSERT INTO
    SPOTIFY_HISTORY(
        user_id,
        timestamp,
        platform,
        ms_played,
        conn_country,
        ip_addr,
        user_agent,
        track_name,
        artist_name,
        album_name,
        spotify_track_uri,
        spotify_artist_uri,
        spotify_album_uri,
        reason_start,
        reason_end,
        shuffle,
        skipped,
        offline,
        offline_timestamp,
        incognito_mode,
        from_history
    )
VALUES
    (
        unnest(@user_ids :: UUID []),
        unnest(@timestamp :: TIMESTAMP []),
        unnest(@platform :: TEXT []),
        unnest(@ms_played :: INTEGER []),
        unnest(@conn_country :: TEXT []),
        unnest(@ip_addr :: TEXT []),
        unnest(@user_agent :: TEXT []),
        unnest(@track_name :: TEXT []),
        unnest(@artist_name :: TEXT []),
        unnest(@album_name :: TEXT []),
        unnest(@spotify_track_uri :: TEXT []),
        unnest(@spotify_artist_uri :: TEXT []),
        unnest(@spotify_album_uri :: TEXT []),
        unnest(@reason_start :: TEXT []),
        unnest(@reason_end :: TEXT []),
        unnest(@shuffle :: BOOLEAN []),
        unnest(@skipped :: BOOLEAN []),
        unnest(@offline :: BOOLEAN []),
        unnest(@offline_timestamp :: TIMESTAMP []),
        unnest(@incognito_mode :: BOOLEAN []),
        unnest(@from_history :: BOOLEAN [])
    ) ON CONFLICT DO NOTHING;

-- name: HistoryGetAll :many
SELECT
    TIMESTAMP,
    TRACK_NAME,
    ARTIST_NAME,
    album_name,
    MS_PLAYED,
    spotify_track_uri,
    spotify_album_uri,
    spotify_artist_uri
FROM
    SPOTIFY_HISTORY
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (
        skipped != true
        OR @include_skips :: boolean
    )
ORDER BY
    timestamp DESC
LIMIT @max_count;

-- name: HistoryGetByTrackURI :many
SELECT
    TIMESTAMP,
    TRACK_NAME,
    ARTIST_NAME,
    album_name,
    MS_PLAYED,
    spotify_track_uri,
    spotify_artist_uri,
    spotify_album_uri
FROM
    SPOTIFY_HISTORY
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (
        skipped != true
        OR @include_skips :: boolean
    )
    AND spotify_track_uri = @uri
ORDER BY
    timestamp ASC;

-- name: HistoryGetByArtistURI :many
SELECT
    TIMESTAMP,
    TRACK_NAME,
    ARTIST_NAME,
    album_name,
    MS_PLAYED,
    spotify_track_uri,
    spotify_artist_uri,
    spotify_album_uri
FROM
    SPOTIFY_HISTORY
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (
        skipped != true
        OR @include_skips :: boolean
    )
    AND spotify_artist_uri = @uri
ORDER BY
    timestamp ASC;

-- name: HistoryGetByAlbumURI :many
SELECT
    TIMESTAMP,
    TRACK_NAME,
    ARTIST_NAME,
    album_name,
    MS_PLAYED,
    spotify_track_uri,
    spotify_artist_uri,
    spotify_album_uri
FROM
    SPOTIFY_HISTORY
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (
        skipped != true
        OR @include_skips :: boolean
    )
    AND spotify_album_uri = @uri
ORDER BY
    timestamp ASC;

-- name: HistoryGetTimestampRange :one
SELECT
    MIN(timestamp) :: timestamp AS first,
    MAX(timestamp) :: timestamp AS last
FROM
    spotify_history
WHERE
    user_id = @user_id;

-- name: HistoryGetArtistStreamCountByYear :many
SELECT
    artist_name,
    COUNT(*) AS occurrences
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (
        skipped != true
        OR @include_skips :: boolean
    )
    AND timestamp BETWEEN (@year :: int || '-01-01 00:00:00') :: timestamp
    AND (
        cast(((@year :: int) + 1) as text) || '-01-01 00:00:00'
    ) :: timestamp
GROUP BY
    artist_name
ORDER BY
    COUNT(*) DESC
LIMIT
    35;

-- name: HistoryGetAlbumStreamCountByYear :many
SELECT
    album_name,
    COUNT(*) AS occurrences
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (
        skipped != true
        OR @include_skips :: boolean
    )
    AND timestamp BETWEEN (@year :: int || '-01-01 00:00:00') :: timestamp
    AND (
        cast(((@year :: int) + 1) as text) || '-01-01 00:00:00'
    ) :: timestamp
GROUP BY
    album_name
ORDER BY
    COUNT(*) DESC
LIMIT
    35;

-- name: HistoryGetTrackStreamCountByYear :many
SELECT
    track_name,
    COUNT(*) AS occurrences
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (
        skipped != true
        OR @include_skips :: boolean
    )
    AND timestamp BETWEEN (@year :: int || '-01-01 00:00:00') :: timestamp
    AND (
        cast(((@year :: int) + 1) as text) || '-01-01 00:00:00'
    ) :: timestamp
GROUP BY
    track_name
ORDER BY
    COUNT(*) DESC
LIMIT
    35;

-- name: HistoryGetTopTracksInTimeframe :many
SELECT
    spotify_track_uri,
    COUNT(*) AS occurrences
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (
        skipped != true
        OR @include_skips :: boolean
    )
    AND timestamp BETWEEN @start_date :: timestamp
    AND @end_date :: timestamp
    AND (
        sqlc.narg(artist_uri)::text IS NULL
        OR spotify_artist_uri = sqlc.narg(artist_uri)::text
    )
    AND (
        sqlc.narg(album_uri)::text IS NULL
        OR spotify_album_uri = sqlc.narg(album_uri)::text
    )
GROUP BY
    spotify_track_uri
ORDER BY
    COUNT(*) DESC
LIMIT
    @max_tracks;

-- name: HistoryGetTopArtistsInTimeframe :many
SELECT
    spotify_artist_uri,
    COUNT(*) AS occurrences,
    string_agg(track_name, '|~|')::text as TRACKS
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (
        skipped != true
        OR @include_skips :: boolean
    )
    AND timestamp BETWEEN @start_date :: timestamp
    AND @end_date :: timestamp
GROUP BY
    spotify_artist_uri
ORDER BY
    COUNT(*) DESC
LIMIT
    @max_tracks;

-- name: HistoryGetTopAlbumsInTimeframe :many
SELECT
    spotify_album_uri,
    COUNT(*) AS occurrences,
    string_agg(track_name, '|~|')::text as TRACKS
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (
        skipped != true
        OR @include_skips :: boolean
    )
    AND timestamp BETWEEN @start_date :: timestamp
    AND @end_date :: timestamp
    AND (
        sqlc.narg(artist_uri)::text IS NULL
        OR spotify_artist_uri = sqlc.narg(artist_uri)::text
    )
GROUP BY
    spotify_album_uri
ORDER BY
    COUNT(*) DESC
LIMIT
    @max_tracks;

-- name: HistoryGetTrackURIForAlbum :one
SELECT spotify_track_uri
FROM
spotify_history
WHERE
artist_name = @artist_name
AND user_id = @user_id
LIMIT 1;

-- name: HistorySetURIsForTrack :exec
UPDATE
spotify_history
SET
    spotify_artist_uri = $2,
    spotify_album_uri = $3
WHERE
    spotify_track_uri = $1;

-- name: HistoryGetTopTracksWithoutURIs :many
SELECT
  SPOTIFY_TRACK_URI,
  COUNT(SPOTIFY_TRACK_URI)
FROM
  spotify_history
WHERE
	spotify_artist_uri IS NULL
GROUP BY
  SPOTIFY_TRACK_URI
ORDER BY
  COUNT DESC
LIMIT
  50;