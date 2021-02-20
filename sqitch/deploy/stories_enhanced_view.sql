-- Deploy scribly:stories_enhanced_view to pg
BEGIN;
CREATE OR REPLACE VIEW stories_enhanced AS
SELECT
    story.*,
    CASE WHEN story.state = 'in_progress' THEN
    (array_agg(cowriters.user_id ORDER BY turn_index)::int[])[((turn_count).count % COUNT(DISTINCT cowriters)) + 1]
ELSE
    NULL
    END AS current_writer_id
FROM
    stories AS story
    CROSS JOIN LATERAL (
        SELECT
            COUNT(*) AS count
        FROM
            turns
        WHERE
            turns.story_id = story.id) turn_count
    LEFT JOIN story_cowriters cowriters ON cowriters.story_id = story.id
GROUP BY
    story.id,
    turn_count.count;
COMMIT;

