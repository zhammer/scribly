-- Verify scribly:scribbot_user on pg

BEGIN;

SELECT * FROM users WHERE username = 'scribbot';

ROLLBACK;
