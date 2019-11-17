Feature: Landing Page
    On the landing page I learn about scribly and can login.

    Scenario: I visit scribly
        When I visit "/"
        Then I see the text "scribly"
        And I see the text "write stories together"
        And I see the text "log in"
        And I see the text "sign up"

    Scenario: I am redirected to me page if I'm already logged in
        Given the following users exist
            | username |
            | zach     |
        And I am logged in as zach
        When I visit "/"
        Then I am on "/me"
