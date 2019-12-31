Feature: Exception Page
    Unhandled exceptions are surfaced to users in a transparent way.

    Scenario: An exception occurs
        # /exception is a route for debugging that raises an exception
        When I visit "/exception" expecting a non-200 response
        Then I see the text "There was an error!"
        And I see the text "You can report this error by"
        And I see the text "sending an email"
        And I see the text "making an issue"
        And I see the text "Exception: Raising an exception, intentionally!"
        And I see the text "raise Exception("Raising an exception, intentionally!")"
