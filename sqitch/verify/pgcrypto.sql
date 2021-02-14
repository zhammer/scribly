-- Verify scribly:pgcrypto on pg
BEGIN;

SELECT
    crypt('random_string', gen_salt('bf', 8));

ROLLBACK;
