CREATE TABLE users (
    id serial PRIMARY KEY,
    login character varying(250) NOT NULL,
    lat double precision,
    long double precision,
    blocked boolean DEFAULT false,
    created timestamp DEFAULT current_timestamp
);

CREATE TABLE messages (
    id serial PRIMARY KEY,
    user_id integer references users(id),
    body text NOT NULL,
    created timestamp DEFAULT current_timestamp
);

CREATE TABLE posts (
    id serial PRIMARY KEY,
    user_id integer references users(id),
    body text NOT NULL,
    created timestamp DEFAULT current_timestamp
);

CREATE TABLE comments (
    id serial PRIMARY KEY,
    user_id integer references users(id),
    post_id integer references posts(id),
    body text NOT NULL,
    created timestamp DEFAULT current_timestamp
);

