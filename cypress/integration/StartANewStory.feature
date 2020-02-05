Feature: Start A New Story

    Background:
        Given the following users exist
            | username   |
            | zhammer    |
            | gsnussbaum |

    Scenario: I need an accessible new story page
        Given I am logged in as zhammer
        When I visit "/new"
        Then I see the text "new story"
        Then the page is accessible

    @focus
    Scenario Outline: I start a new story
        Given I am logged in as zhammer
        When I visit "/new"
        And I click on the "title" input
        And I type "<title>"
        And I click on the "body" textarea
        And I type "<intro>"
        And I click the text "add cowriters"
        Then I am on "/stories/1"
        And I see the text "cowriters"
        And I see the text "<title>"
        And I see the text "<intro>"
        And the page is accessible

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
        And I see the text "cowriters"
        And I see the text "submit"
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

    Scenario: Gabe looks at our story after I add him
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
        And I log in as "gsnussbaum"
        And I visit "/stories/1"
        Then I see the text "my title"
        And I see the text "write"
        And I see the text "write and finish"
        And I see the text "finish"
        And I see the text "pass"
