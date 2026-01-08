Feature: Exception Page
    Unhandled exceptions are surfaced to users in a transparent way.

    Scenario: An exception occurs
        # /exception is a route for debugging that raises an exception
        When I visit "/exception" expecting a non-200 response
        Then I see the text "There was an error!"
        And I see the text "Here is an error that was raised while you were using Scribly."
        And I see the text "send an email"
        And I see the text "make an issue and/or write some code"
        And I see the text "*errors.errorString: Raising an exception, intentionally!"
        And the page is accessible