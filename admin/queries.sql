-- name: TableSizesAndRows :many
SELECT
  nspname AS schema,
  relname AS table,
  reltuples AS rows_estimate,
  pg_relation_size(c.oid) AS rel_size_bytes,
  pg_indexes_size(c.oid) AS index_size_bytes,
  pg_total_relation_size(c.oid) AS total_size_bytes
FROM
  pg_catalog.pg_class c
  JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE
  relkind = 'r'
  AND nspname = 'public'
ORDER BY
  total_size_bytes DESC;

-- name: UncachedTracks :many
SELECT
  SPOTIFY_TRACK_URI,
  COUNT(SPOTIFY_TRACK_URI),
  TRACK_NAME,
  h.artist_name
FROM
  spotify_history h
  LEFT JOIN spotify_track_cache tc ON h.spotify_track_uri = tc.uri
WHERE
  tc.uri IS NULL
GROUP BY
  SPOTIFY_TRACK_URI,
  TRACK_NAME,
  h.ARTIST_NAME
ORDER BY
  COUNT DESC;

-- name: MissingISRCNumbers :many
SELECT
  h.*,
  tc.isrc
FROM
  SPOTIFY_HISTORY h
  JOIN spotify_track_cache tc ON h.ISRC IS NULL
    AND tc.isrc IS NOT NULL
    AND tc.uri = h.spotify_track_uri
  ORDER BY
    TIMESTAMP DESC;

-- name: MissingArtistURIs :many
SELECT
  SPOTIFY_TRACK_URI,
  COUNT(SPOTIFY_TRACK_URI),
(array_agg(track_name))[1] AS track_name,
(array_agg(artist_name))[1] AS artist_name,
(array_agg(DISTINCT user_id))::text[] AS user_ids
FROM
  spotify_history
WHERE
  spotify_artist_uri IS NULL
GROUP BY
  SPOTIFY_TRACK_URI
ORDER BY
  COUNT DESC;

