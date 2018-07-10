-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE Users (
  id TEXT,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE Rooms (
  id TEXT,
  display_name TEXT NOT NULL,
  normalized_name TEXT NOT NULL,
  rotator BLOB NOT NULL,
	rotator_type INTEGER NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE Tracks (
  id TEXT,
  track BLOB NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE QueueTracks (
  id TEXT NOT NULL,
  previous_id TEXT,
  next_id TEXT,
  track_id NOT NULL,
	room_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	added_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  played BOOLEAN NOT NULL CHECK (played IN (0,1)),
	FOREIGN KEY (previous_id) REFERENCES QueueTracks(id),
	FOREIGN KEY (next_id) REFERENCES QueueTracks(id),
	FOREIGN KEY (track_id) REFERENCES Tracks(id),
	FOREIGN KEY (room_id) REFERENCES Rooms(id),
	FOREIGN KEY (user_id) REFERENCES Users(id),
	PRIMARY KEY (id)
);

CREATE TABLE Queues (
	room_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	next_queue_track_id TEXT,
	joined_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
	FOREIGN KEY (next_queue_track_id) REFERENCES QueueTracks(id),
	FOREIGN KEY (room_id) REFERENCES Rooms(id),
	FOREIGN KEY (user_id) REFERENCES Users(id),
	PRIMARY KEY (room_id, user_id)
);

CREATE TABLE History (
	room_id TEXT,
	track_entries BLOB NOT NULL,
	FOREIGN KEY (room_id) REFERENCES Rooms(id)
	PRIMARY KEY (room_id)
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE History;
DROP TABLE Queues;
DROP TABLE QueueTracks;
DROP TABLE Tracks;
DROP TABLE Rooms;
