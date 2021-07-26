-- Deploy scribly:scribbot_user to pg

BEGIN;

INSERT INTO users (username, password, email) VALUES
    ('scribbot', '', '')
;

COMMIT;
