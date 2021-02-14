-- Verify scribly:useless_table on pg
BEGIN;

SELECT
    id,
    note
FROM
    useless_table
WHERE
    FALSE;

ROLLBACK;
