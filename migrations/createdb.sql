CREATE TABLE IF NOT EXISTS users (
    id SERIAL,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO users (username, password)
VALUES
    ('zach.the.hammer@gmail.com', 'password'),
    ('gsnussbaum@gmail.com', 'password')
;


