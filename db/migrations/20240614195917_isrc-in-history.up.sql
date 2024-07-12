ALTER TABLE spotify_history
  ADD COLUMN isrc text;

CREATE INDEX history_isrc_idx ON spotify_history(isrc);

CREATE INDEX history_uri_idx ON spotify_history(spotify_track_uri);

CREATE INDEX track_cache_isrc_idx ON spotify_track_cache(isrc);

CREATE INDEX track_cache_uri_idx ON spotify_track_cache(uri);

