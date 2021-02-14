-- Revert scribly:useless_table from pg
BEGIN;

DROP TABLE useless_table;

COMMIT;
