Feature: Landing Page
    On the landing page I learn about scribly and can login.

    Scenario: I visit scribly
        When I visit "/"
        Then I see the text "scribly"
        And I see the text "write stories together"
        And I see the text "log in"
        And I see the text "sign up"
