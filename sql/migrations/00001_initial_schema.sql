-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE Users (
  id TEXT PRIMARY KEY,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL
);

CREATE TABLE Rooms (
  id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  rotator BLOB NOT NULL,
	rotator_type INTEGER NOT NULL,
	music_service INTEGER NOT NULL
);

CREATE TABLE Queues (
	room_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	offset INTEGER DEFAULT 0,
	tracks BLOB NOT NULL,
	joined_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
	FOREIGN KEY (room_id) REFERENCES Rooms(id),
	FOREIGN KEY (user_id) REFERENCES Users(id),
	PRIMARY KEY (room_id, user_id)
);

CREATE TABLE History (
	room_id TEXT PRIMARY KEY,
	track_entries BLOB NOT NULL,
	FOREIGN KEY (room_id) REFERENCES Rooms(id)
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE History;
DROP TABLE Queues;
DROP TABLE Rooms;
DROP TABLE Users;
