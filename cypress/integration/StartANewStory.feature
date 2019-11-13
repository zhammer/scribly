Feature: Start A New Story

    Background:
        Given the following users exist
            | username | password |
            | zhammer  | password |

    Scenario Outline: I start a new story
        Given I am logged in as zhammer
        When I visit "/new"
        And I click on the "title" input
        And I type "<title>"
        And I click on the "intro" textarea
        And I type "<intro>"
        And I click the text "add cowriters"
        Then I see the text "add cowriters"
        And I see the text "<title>"
        And I see the text "<intro>"

        Examples:
            | title                    | intro                                                        |
            | The old man and the seed | After a long life of fishing, the old man took up gardening. |
