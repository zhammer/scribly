Feature: Scribbot
    Scribbot uses openai to take turns writing in stories.

    Scenario: It's scribbot's turn
        Given the following users exist
            | username |
            | scribbot |
            | zach     |
        And the follow stories exist
            | title                | turns | users          |
            | scribbot's big break | 1     | zach, scribbot |
        And I am logged in as zach
        When I wait for scribbot to take its turn
        And I visit "/stories/1"
        Then I see the text "He could sit and contemplate or he could change the world."

    Scenario: It is not scribbot's turn
        Given the following users exist
            | username |
            | scribbot |
            | zach     |
        And the follow stories exist
            | title                | turns | users          |
            | scribbot's big break | 2     | zach, scribbot |
        And I am logged in as zach
        When I wait for scribbot to take its turn
        And I visit "/stories/1"
        Then I do not see the text "He could sit and contemplate or he could change the world."