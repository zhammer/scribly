Feature: Me Page
    On the Me Page I can see my stories and I can start a new story.

    Background:
        Given the following users exist
            | username | password |
            | zhammer  | password |

    Scenario: I visit the Me Page
        Given I am logged in as zhammer:password
        When I visit "/me"
        Then I see the text "start a new story"

    Scenario: I try to visit the Me Page without being logged in
        Given I am not logged in
        When I visit "/me"
        Then I am on "/"
        And I see the text "log in"

    Scenario: I click start a new story
        Given I am logged in as zhammer:password
        When I visit "/me"
        And I click the text "start a new story"
        Then I am on "/new"
        And I see the text "start a new story"
