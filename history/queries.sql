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
        reason_start,
        reason_end,
        shuffle,
        skipped,
        offline,
        offline_timestamp,
        incognito_mode
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
        @reason_start,
        @reason_end,
        @shuffle,
        @skipped,
        @offline,
        @offline_timestamp,
        @incognito_mode
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
        reason_start,
        reason_end,
        shuffle,
        skipped,
        offline,
        offline_timestamp,
        incognito_mode
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
        unnest(@reason_start :: TEXT []),
        unnest(@reason_end :: TEXT []),
        unnest(@shuffle :: BOOLEAN []),
        unnest(@skipped :: BOOLEAN []),
        unnest(@offline :: BOOLEAN []),
        unnest(@offline_timestamp :: TIMESTAMP []),
        unnest(@incognito_mode :: BOOLEAN [])
    ) ON CONFLICT DO NOTHING;

-- name: HistoryGetAll :many
SELECT
    TIMESTAMP,
    TRACK_NAME,
    ARTIST_NAME,
    album_name,
    MS_PLAYED,
    spotify_track_uri
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
    timestamp ASC;

-- name: HistoryGetByURI :many
SELECT
    TIMESTAMP,
    TRACK_NAME,
    ARTIST_NAME,
    album_name,
    MS_PLAYED,
    spotify_track_uri
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

-- name: HistoryGetTimestampRange :one
SELECT
    MIN(timestamp) :: timestamp AS first,
    MAX(timestamp) :: timestamp AS last
FROM
    spotify_history
WHERE
    user_id = @user_id;

-- name: HistoryGetArtistStreamsByYear :many
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

-- name: HistoryGetAlbumStreamsByYear :many
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

-- name: HistoryGetTrackStreamsByYear :many
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