-- Deploy scribly:pgcrypto to pg
BEGIN;

CREATE EXTENSION pgcrypto;

COMMIT;
