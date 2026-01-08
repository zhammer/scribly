-- Verify scribly:remove_scribbot_user on pg

BEGIN;

SELECT 1 WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'scribbot');

ROLLBACK;
