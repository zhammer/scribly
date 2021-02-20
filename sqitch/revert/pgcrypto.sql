-- Revert scribly:pgcrypto from pg
BEGIN;

DROP EXTENSION pgcrypto;

COMMIT;
