Feature: Me Page
    On the Me Page I can see my stories and I can start a new story.

    Background:
        Given the following users exist
            | username |
            | zach     |
            | gabe     |

    Scenario Outline: I visit the Me Page as <username>
        Given I am logged in as <username>
        When I visit "/me"
        Then I see the text "<username>'s scribly"
        And I see the text "start a new story"

        Examples:
            | username |
            | zach     |
            | gabe     |


    Scenario: I try to visit the Me Page without being logged in
        Given I am not logged in
        When I visit "/me"
        Then I am on "/"
        And I see the text "log in"

    Scenario: I click start a new story
        Given I am logged in as zach
        When I visit "/me"
        And I click the text "start a new story"
        Then I am on "/new"
        And I see the text "add cowriters (saves current draft)"

    @focus
    Scenario: I have several stories on my account
        Given the following stories exist
            | title        | turns | users      | complete |
            | The cool dog | 8     | zach, gabe | false    |
        And I am logged in as zach
        When I visit "/me"
