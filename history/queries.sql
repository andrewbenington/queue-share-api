-- name: HistoryInsertOne :exec
INSERT INTO SPOTIFY_HISTORY(
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
    from_history,
    isrc)
VALUES (
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
    @from_history,
    @isrc);

-- name: HistoryInsertBulk :exec
INSERT INTO SPOTIFY_HISTORY(
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
    from_history,
    isrc)
VALUES (
    unnest(
        @user_ids::uuid[]),
    unnest(
        @timestamp::timestamp[]),
    unnest(
        @platform::text[]),
    unnest(
        @ms_played::integer[]),
    unnest(
        @conn_country::text[]),
    unnest(
        @ip_addr::text[]),
    unnest(
        @user_agent::text[]),
    unnest(
        @track_name::text[]),
    unnest(
        @artist_name::text[]),
    unnest(
        @album_name::text[]),
    unnest(
        @spotify_track_uri::text[]),
    unnest(
        @spotify_artist_uri::text[]),
    unnest(
        @spotify_album_uri::text[]),
    unnest(
        @reason_start::text[]),
    unnest(
        @reason_end::text[]),
    unnest(
        @shuffle::boolean[]),
    unnest(
        @skipped::boolean[]),
    unnest(
        @offline::boolean[]),
    unnest(
        @offline_timestamp::timestamp[]),
    unnest(
        @incognito_mode::boolean[]),
    unnest(
        @from_history::boolean[]),
    unnest(
        @isrc::text[]))
ON CONFLICT
    DO NOTHING;

-- name: HistoryGetAll :many
SELECT
    TIMESTAMP,
    TRACK_NAME,
    h.ARTIST_NAME,
    h.album_name,
    MS_PLAYED,
    spotify_track_uri,
    spotify_album_uri,
    spotify_artist_uri,
    image_url,
    other_artists,
    h.isrc
FROM
    SPOTIFY_HISTORY h
    JOIN SPOTIFY_TRACK_CACHE ON URI = spotify_track_uri
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND (sqlc.narg(start_date)::timestamp IS NULL
        OR sqlc.narg(end_date)::timestamp IS NULL
        OR timestamp BETWEEN sqlc.narg(start_date)::timestamp AND sqlc.narg(end_date)::timestamp)
ORDER BY
    timestamp DESC
LIMIT @max_count;

-- name: HistoryGetByTrackURI :many
SELECT
    TIMESTAMP,
    h.TRACK_NAME,
    h.ARTIST_NAME,
    h.album_name,
    MS_PLAYED,
    spotify_track_uri,
    spotify_artist_uri,
    spotify_album_uri,
    tc1.isrc
FROM
    SPOTIFY_HISTORY h
    JOIN spotify_track_cache tc1 ON tc1.uri = @uri
    JOIN spotify_track_cache tc2 ON tc2.isrc = tc1.isrc
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND h.spotify_track_uri = tc2.uri
ORDER BY
    timestamp ASC;

-- name: HistoryGetByArtistURI :many
SELECT
    sh.TIMESTAMP,
    sh.TRACK_NAME,
    sh.album_name,
    sh.MS_PLAYED,
    sh.spotify_track_uri,
    sh.spotify_album_uri,
    tc.isrc
FROM
    SPOTIFY_HISTORY sh
    JOIN SPOTIFY_TRACK_CACHE tc ON sh.spotify_track_uri = tc.uri
        AND user_id = @user_id
        AND ms_played >= @min_ms_played
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
    AND spotify_album_uri = @uri
ORDER BY
    timestamp ASC;

-- name: HistoryGetTimestampRange :one
SELECT
    MIN(timestamp)::timestamp AS first,
    MAX(timestamp)::timestamp AS last
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
    AND timestamp BETWEEN (@year::int || '-01-01 00:00:00')::timestamp AND (cast(((@year::int) + 1) AS text) || '-01-01 00:00:00')::timestamp
GROUP BY
    artist_name
ORDER BY
    COUNT(*) DESC
LIMIT 35;

-- name: HistoryGetAlbumStreamCountByYear :many
SELECT
    album_name,
    COUNT(*) AS occurrences
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND timestamp BETWEEN (@year::int || '-01-01 00:00:00')::timestamp AND (cast(((@year::int) + 1) AS text) || '-01-01 00:00:00')::timestamp
GROUP BY
    album_name
ORDER BY
    COUNT(*) DESC
LIMIT 35;

-- name: HistoryGetTrackStreamCountByYear :many
SELECT
    track_name,
    COUNT(*) AS occurrences
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND timestamp BETWEEN (@year::int || '-01-01 00:00:00')::timestamp AND (cast(((@year::int) + 1) AS text) || '-01-01 00:00:00')::timestamp
GROUP BY
    track_name
ORDER BY
    COUNT(*) DESC
LIMIT 35;

-- name: HistoryGetTopTracksInTimeframe :many
SELECT
    spotify_track_uri,
    COUNT(*) AS occurrences
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND timestamp BETWEEN @start_date::timestamp AND @end_date::timestamp
    AND (sqlc.narg(artist_uris)::text[] IS NULL
        OR spotify_artist_uri = ANY (sqlc.narg(artist_uris)::text[]))
    AND (sqlc.narg(album_uri)::text IS NULL
        OR spotify_album_uri = sqlc.narg(album_uri)::text)
GROUP BY
    spotify_track_uri
ORDER BY
    COUNT(*) DESC
LIMIT @max_tracks;

-- name: HistoryGetTopTracksInTimeframeDedup :many
WITH top_isrcs AS (
    SELECT
        tc.isrc,
        COUNT(*) AS occurrences,
(json_agg(DISTINCT h.spotify_track_uri)) AS spotify_track_uris
    FROM
        spotify_history h
        JOIN spotify_track_cache tc ON tc.uri = h.spotify_track_uri
    WHERE
        user_id = @user_id
        AND ms_played >= @min_ms_played
        AND timestamp BETWEEN @start_date::timestamp AND @end_date::timestamp
        AND (sqlc.narg(artist_uris)::text[] IS NULL
            OR spotify_artist_uri = ANY (sqlc.narg(artist_uris)::text[]))
        AND (sqlc.narg(album_uri)::text IS NULL
            OR h.spotify_album_uri = sqlc.narg(album_uri)::text)
    GROUP BY
        tc.isrc
    ORDER BY
        COUNT(*) DESC
    LIMIT @max_tracks
),
pref_albums AS (
    SELECT DISTINCT ON (top_isrcs.isrc)
        top_isrcs.*,
        tc.uri AS spotify_track_uri
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
        release_date ASC
)
SELECT
    *
FROM
    pref_albums
ORDER BY
    occurrences DESC;

-- name: HistoryGetTopArtistsInTimeframe :many
SELECT
    spotify_artist_uri,
    COUNT(*) AS occurrences,
    json_agg(spotify_track_uri) AS TRACKS
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND timestamp BETWEEN @start_date::timestamp AND @end_date::timestamp
GROUP BY
    spotify_artist_uri
ORDER BY
    COUNT(*) DESC
LIMIT @max;

-- name: HistoryGetTopAlbumsInTimeframe :many
SELECT
    spotify_album_uri,
    COUNT(*) AS occurrences,
    json_agg(spotify_track_uri) AS TRACKS
FROM
    spotify_history
WHERE
    user_id = @user_id
    AND ms_played >= @min_ms_played
    AND timestamp BETWEEN @start_date::timestamp AND @end_date::timestamp
    AND (sqlc.narg(artist_uri)::text IS NULL
        OR spotify_artist_uri = sqlc.narg(artist_uri)::text)
GROUP BY
    spotify_album_uri
ORDER BY
    COUNT(*) DESC
LIMIT @max;

-- name: HistoryGetTrackURIForAlbum :one
SELECT
    spotify_track_uri
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
    spotify_artist_uri = @spotify_artist_uri,
    spotify_album_uri = @spotify_album_uri
WHERE
    spotify_track_uri = @spotify_track_uri;

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
LIMIT 50;

-- name: HistoryGetTopTracksNotInCache :many
SELECT
    SPOTIFY_TRACK_URI,
    COUNT(SPOTIFY_TRACK_URI)
FROM
    spotify_history h
    LEFT JOIN spotify_track_cache tc ON h.spotify_track_uri = tc.uri
WHERE
    tc.uri IS NULL
GROUP BY
    SPOTIFY_TRACK_URI
ORDER BY
    COUNT DESC
LIMIT 50;

-- name: HistoryGetTopAlbumsNotInCache :many
SELECT
    SPOTIFY_ALBUM_URI,
    COUNT(SPOTIFY_ALBUM_URI)
FROM
    spotify_history h
    LEFT JOIN spotify_album_cache ac ON h.spotify_album_uri = ac.uri
WHERE
    ac.uri IS NULL
GROUP BY
    SPOTIFY_ALBUM_URI
ORDER BY
    COUNT DESC
LIMIT 50;

-- name: HistoryGetRecentArtistStreams :many
WITH SongStreamCounts AS (
    SELECT
        spotify_artist_uri,
        spotify_track_uri,
        COUNT(*) AS stream_count
    FROM
        spotify_history
    WHERE
        user_id = @user_id
        AND spotify_artist_uri = ANY (@artist_uris::text[])
        AND timestamp >= CURRENT_DATE - INTERVAL '18 months'
    GROUP BY
        spotify_artist_uri,
        spotify_track_uri
),
RankedStreams AS (
    SELECT
        spotify_artist_uri,
        spotify_track_uri,
        stream_count,
        ROW_NUMBER() OVER (PARTITION BY spotify_artist_uri ORDER BY stream_count DESC) AS rank
    FROM
        SongStreamCounts
)
SELECT
    spotify_artist_uri,
    spotify_track_uri,
    stream_count
FROM
    RankedStreams
WHERE
    rank <= 10
ORDER BY
    spotify_artist_uri,
    rank;

-- name: HistoryGetNewArtists :many
WITH all_artists AS (
    SELECT
        spotify_artist_uri,
        count(*),
        array_agg(DISTINCT date(timestamp))::text[] AS distinct_dates
    FROM
        spotify_history h1
    WHERE
        h1.user_id = @user_id
        AND h1.timestamp >= @start_date::timestamp
        AND h1.timestamp < @end_date::timestamp
    GROUP BY
        spotify_artist_uri
),
new_this_month AS (
    SELECT
        *
    FROM
        all_artists aa
    WHERE
        NOT EXISTS (
            SELECT
                1
            FROM
                spotify_history h2
            WHERE
                h2.user_id = @user_id
                AND timestamp < @start_date::timestamp
                AND h2.spotify_artist_uri = aa.spotify_artist_uri))
SELECT
    TRIM(LEADING 'spotify:artist:' FROM spotify_artist_uri)::text AS id,
    count,
    distinct_dates
FROM
    new_this_month
WHERE
    count > 5
    AND array_length(distinct_dates, 1) > 1
LIMIT 100;

-- name: HistoryGetTopAlbumsByRelease :many
WITH history AS (
    SELECT
        sh.spotify_album_uri,
        date_trunc(@interval, ac.release_date)::date AS release_interval
    FROM
        spotify_history sh
        JOIN spotify_album_cache ac ON ac.uri = sh.spotify_album_uri
    WHERE
        sh.user_id = @user_id
        AND ac.album_type = 'album'
),
grouped_history AS (
    SELECT DISTINCT ON (release_interval)
        *,
        count(*)
    FROM
        history
    GROUP BY
        spotify_album_uri,
        release_interval
    ORDER BY
        release_interval,
        count DESC
)
SELECT
    release_interval,
    count AS streams,
    name,
    artist_name
FROM
    grouped_history h
    JOIN spotify_album_cache ac ON ac.uri = h.spotify_album_uri
ORDER BY
    release_interval DESC,
    count DESC
LIMIT 1000;

