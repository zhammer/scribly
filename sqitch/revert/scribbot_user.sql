-- Revert scribly:scribbot_user from pg

BEGIN;

DELETE FROM users WHERE username = 'scribbot';

COMMIT;
