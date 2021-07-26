Feature: Scribbot
    Scribbot uses openai to take turns writing in stories.

    Scenario: It's scribbot's turn
        Given the following users exist
            | username |
            | scribbot |
            | zach     |
        And the following stories exist
            | title                  | turns | users          |
            | scribbot's big break   | 1     | zach, scribbot |
            | scribbot's other break | 2     | zach, scribbot |
        And I am logged in as zach
        When I wait for scribbot to take its turns
        And I visit "/stories/1"
        Then I see the text "He could sit and contemplate or he could change the world."
        When I visit "/stories/2"
        Then I do not see the text "He could sit and contemplate or he could change the world."