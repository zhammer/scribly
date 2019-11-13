Feature: Start A New Story

    Background:
        Given the following users exist
            | username   |
            | zhammer    |
            | gsnussbaum |

    Scenario Outline: I start a new story
        Given I am logged in as zhammer
        When I visit "/new"
        And I click on the "title" input
        And I type "<title>"
        And I click on the "intro" textarea
        And I type "<intro>"
        And I click the text "add cowriters"
        Then I am on "/stories/1"
        And I see the text "add cowriters"
        And I see the text "<title>"
        And I see the text "<intro>"

        Examples:
            | title                    | intro                                                        |
            | The old man and the seed | After a long life of fishing, the old man took up gardening. |

    Scenario: My draft saves upon clicking "add cowriters"
        Given I am logged in as zhammer
        When I visit "/new"
        And I click on the "title" input
        And I type "my title"
        And I hit tab
        And I type "my intro"
        And I click the text "add cowriters"
        And I refresh the page
        Then I am on "/stories/1"
        And I see the text "add cowriters"
        And I see the text "my title"
        And I see the text "my intro"

    Scenario: I add cowriters to my story
        Given I am logged in as zhammer
        When I visit "/new"
        And I click on the "title" input
        And I type "my title"
        And I hit tab
        And I type "my intro"
        And I click the text "add cowriters"
        And I click on the "person-1" input
        And I type "gsnussbaum"
        And I click the text "submit"
        Then I am on "/stories/1"
        And I see the text "my title"
        And I see the text "cowriters: zhammer, gsnussbaum"
        And I see the text "turn: 1 (gsnussbaum's turn)"
        And I see the text "my intro"
