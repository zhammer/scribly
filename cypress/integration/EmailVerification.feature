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

    @focus
    Scenario: I click the resend verification link button
        Given I am logged in as zach
        When I visit "/me"
        And I click the button "resend verification link"
        Then I see the text "Email verification sent!"
        And I see the text "You should receive an email at zach@mail.com with a verification link shortly!"
        And I received an email at "zach@mail.com" with the subject "Verify your email"
