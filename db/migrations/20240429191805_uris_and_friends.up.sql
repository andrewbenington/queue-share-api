ALTER TABLE spotify_history
  ADD COLUMN spotify_artist_uri TEXT,
  ADD COLUMN spotify_album_uri TEXT,
  ADD COLUMN from_history BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE user_friends(
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  friend_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  added_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT user_friends_pkey PRIMARY KEY (user_id, friend_id)
);

CREATE TABLE user_friend_requests(
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  friend_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  request_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT user_friend_requests_pkey PRIMARY KEY (user_id, friend_id)
  );