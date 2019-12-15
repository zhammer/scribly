Feature: Email Verification

    Background:
        Given the following users exist
            | username |
            | zach     |

    Scenario: I am prompted to verify my email on the me page
        Given I am logged in as zach
        When I visit "/me"
        Then I see the text "verify your email to follow stories you're working on!"
        And I see the button "resend verification link"
