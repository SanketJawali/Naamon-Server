CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);

CREATE TABLE api_maps (
    id INTEGER PRIMARY KEY,
    key TEXT UNIQUE NOT NULL,
    target_url TEXT
);
