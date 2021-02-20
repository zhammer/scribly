-- Verify scribly:stories_enhanced_view on pg
BEGIN;
SELECT
    id,
    title,
    current_writer_id
FROM
    stories_enhanced
WHERE
    FALSE;
ROLLBACK;

