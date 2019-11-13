CREATE TABLE IF NOT EXISTS users (
    id SERIAL,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

CREATE TYPE story_state AS ENUM ('draft', 'in_progress', 'done');

CREATE TABLE IF NOT EXISTS stories (
    id SERIAL,
    title TEXT NOT NULL,
    state story_state NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS story_cowriters (
    story_id INTEGER NOT NULL REFERENCES stories (id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    turn_index INTEGER NOT NULL,
    PRIMARY KEY (story_id, user_id),
    UNIQUE (story_id, turn_index)
);

CREATE TYPE turn_action AS ENUM ('pass', 'write', 'finish', 'write_and_finish');

CREATE TABLE IF NOT EXISTS turns (
    id SERIAL,
    story_id INTEGER NOT NULL REFERENCES stories (id),
    taken_by INTEGER NOT NULL REFERENCES users (id),
    action turn_action NOT NULL,
    text_written text,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);
