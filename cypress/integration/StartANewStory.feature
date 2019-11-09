Feature: Start A New Story

    Scenario Outline: I start a new story
        Given I am logged in
        When I visit "/new"
        And I click on the "title" input
        And I type "<title>"
        And I click on the "intro" textarea
        And I type "<intro>"

        Examples:
            | title                    | intro                                                        |
            | The old man and the seed | After a long life of fishing, the old man took up gardening. |


