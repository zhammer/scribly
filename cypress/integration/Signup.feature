Feature: Sign up
    I can sign up for a scribly account!

    Scenario: I sign up for scribly
        When I visit "/"
        And I click the text "sign up"
        And I click on the "username" input
        And I type "zach"
        And I click on the "email" input
        And I type "zach@mail.com"
        And I click on the "password" input
        And I type "mypassword"
        And I click on the "password_confirmation" input
        And I type "mypassword"
        And I click the button "sign up"
        Then I am on "/me"
        And I see the text "zach's scribly"
        And I received an email at "zach@mail.com" with the subject "Verify your email"
