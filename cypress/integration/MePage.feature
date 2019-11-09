Feature: Me Page
    On the Me Page I can see my stories and I can start a new story.

    Scenario: I visit the Me Page
        Given I am logged in
        When I visit "/me"
        Then I see the text "start a new story"

    Scenario: I try to visit the Me Page without being logged in
        Given I am not logged in
        When I visit "/me"
        Then I am on "/"
        And I see the text "log in"
