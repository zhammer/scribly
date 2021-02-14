-- Deploy scribly:useless_table to pg
BEGIN;

CREATE TABLE useless_table (
    id SERIAL,
    note TEXT,
    PRIMARY KEY (id)
);

COMMIT;
