-- Revert scribly:remove_scribbot_user from pg

BEGIN;

INSERT INTO users (username, password, email) VALUES
    ('scribbot', '', '')
;

COMMIT;
