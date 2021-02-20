-- Revert scribly:user_story_view from pg
BEGIN;

DROP VIEW user_stories;

COMMIT;
