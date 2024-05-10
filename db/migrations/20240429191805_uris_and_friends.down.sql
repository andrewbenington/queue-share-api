ALTER TABLE spotify_history
  DROP COLUMN spotify_artist_uri,
  DROP COLUMN spotify_album_uri,
  DROP COLUMN from_history;

DROP TABLE IF EXISTS user_friends;
DROP TABLE IF EXISTS user_friend_requests;