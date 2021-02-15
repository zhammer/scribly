-- Revert scribly:stories_enhanced_view from pg
BEGIN;
DROP VIEW stories_enhanced;
COMMIT;

