-- Deploy scribly:user_story_view to pg
BEGIN;

CREATE
OR REPLACE VIEW user_stories AS
SELECT
    users.id AS user_id,
    stories.id AS story_id,
    (
        CASE
            WHEN user_story_hides.hidden_status = 'hidden' THEN true
            ELSE false
        END
    ) AS hidden
FROM
    stories
    LEFT JOIN story_cowriters ON story_cowriters.story_id = stories.id
    JOIN users ON (
        users.id = story_cowriters.user_id
        OR users.id = stories.created_by
    )
    LEFT JOIN user_story_hides ON (
        user_story_hides.story_id = stories.id
        AND user_story_hides.user_id = users.id
    )
GROUP BY
    users.id,
    stories.id,
    hidden;

COMMIT;
