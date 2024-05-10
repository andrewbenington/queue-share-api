
CREATE TABLE spotify_track_cache(
  id text PRIMARY KEY,
  uri text NOT NULL,
  name text NOT NULL,
  album_id text NOT NULL,
  album_uri text NOT NULL,
  album_name text NOT NULL,
  artist_id text NOT NULL,
  artist_uri text NOT NULL,
  artist_name text NOT NULL,
  image_url text,
  other_artists jsonb,
  duration_ms int NOT NULL,
  popularity int NOT NULL DEFAULT 0,
  explicit BOOLEAN NOT NULL DEFAULT false,
  preview_url text NOT NULL,
  disc_number int NOT NULL,
  track_number int NOT NULL,
  type text NOT NULL,
  external_ids jsonb,
  isrc text
);

CREATE TABLE spotify_album_cache(
  id text PRIMARY KEY,
  uri text NOT NULL,
  name text NOT NULL,
  artist_id text NOT NULL,
  artist_uri text NOT NULL,
  artist_name text NOT NULL,
  album_group text,
  album_type text,
  image_url text,
  release_date date,
  release_date_precision text,
  genres jsonb,
  popularity int
);


CREATE TABLE spotify_artist_cache(
  id text PRIMARY KEY,
  uri text NOT NULL,
  name text NOT NULL,
  image_url text,
  genres jsonb,
  popularity int,
  follower_count int
);
