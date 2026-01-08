Feature: Log in
    I can log in to scribly!

    Background:
        Given the following users exist
            | username   |
            | zhammer    |
            | gsnussbaum |

    Scenario Outline: I log in to the existing account <username>
        When I visit "/"
        And I click the text "log in"
        And I click on the "username" input
        And I type "<username>"
        And I click on the "password" input
        And I type "password"
        And I click the button "log in"
        Then I am on "/me"
        And I see the text "<username>'s scribly"

        Examples:
            | username   |
            | zhammer    |
            | gsnussbaum |

    Scenario: I need an accessible login page
        When I visit "/login"
        Then the page is accessible
