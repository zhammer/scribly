CREATE TYPE hidden_status_type AS ENUM ('hidden', 'unhidden');

CREATE TABLE IF NOT EXISTS user_story_hides (
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    story_id INTEGER NOT NULL REFERENCES stories (id) ON DELETE CASCADE,
    hidden_status hidden_status_type NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, story_id)
);
