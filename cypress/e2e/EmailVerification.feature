Feature: Email Verification

    Background:
        Given the following users exist
            | username | email_verification_status |
            | zach     | pending                   |

    Scenario: I am prompted to verify my email on the me page
        Given I am logged in as zach
        When I visit "/me"
        Then I see the text "verify your email to follow stories you're working on!"
        And I see the button "resend verification link"

    Scenario: I click the resend verification link button
        Given I am logged in as zach
        When I visit "/me"
        And I click the button "resend verification link"
        Then I see the text "Email verification sent!"
        And I see the text "You should receive an email at zach@mail.com with a verification link shortly!"
        And the page is accessible
        And I received an email at "zach@mail.com" with the subject "Verify your email"

    Scenario: I verify my email
        Given I am logged in as zach
        When I visit "/me"
        And I click the button "resend verification link"
        And I open my email at "zach@mail.com" with the subject "Verify your email"
        And I click the link "verify your email"
        Then I see the text "Your email zach@mail.com has been verified"
        And the page is accessible

    Scenario: My email verification status is reflected on the me page
        Given I am logged in as zach
        When I visit "/me"
        And I click the button "resend verification link"
        And I open my email at "zach@mail.com" with the subject "Verify your email"
        And I click the link "verify your email"
        And I click the link "click this link to go back to scribly"
        Then I do not see the text "verify your email"
