-- Deploy scribly:remove_scribbot_user to pg

BEGIN;

DELETE FROM users WHERE username = 'scribbot';

COMMIT;
