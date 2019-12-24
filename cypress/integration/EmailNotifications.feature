Feature: Email Notifications
    Scribly users get email notifications to keep them updated on their
    stories.

    Background:
        Given the following users exist
            | username | email_verification_status |
            | zach     | verified                  |
            | gabe     | verified                  |
            | rakesh   | verified                  |

    Scenario Outline: Email notification is sent to <recipient> when rakesh takes turn <turnAction>
        Given the following stories exist
            | title       | turns | users              | complete |
            | Black Truck | 5     | zach, gabe, rakesh | false    |
        And I am logged in as rakesh
        When I visit "/stories/1"
        And I click on the "text" textarea
        # turn text is ignored on "pass" or "finish" actions
        And I type "I've been through the fires. I've felt embers down my spine."
        And I click the button "^\W*<turnAction>\W*$"
        And I wait .5 seconds
        And I log in as "<recipient>"
        And I open my email at "<recipient>@mail.com" with the subject "<subject>"
        Then I see the text "<turnInfoText>"
        And I <doOrDoNotSeeWhoseTurnText> see the text "<whoseTurnText>"
        And I <doOrDoNotSeeTurnText> see the text "I've been through the fires. I've felt embers down my spine."

        Examples:
            | turnAction       | recipient | subject                                | turnInfoText                                   | doOrDoNotSeeWhoseTurnText | whoseTurnText    | doOrDoNotSeeTurnText |
            | write            | zach      | It's your turn on Black Truck!         | rakesh wrote a section!                        | do                        | it's your turn   | do                   |
            | write            | gabe      | rakesh took their turn on Black Truck! | rakesh wrote a section!                        | do                        | it's zach's turn | do                   |
            | write and finish | zach      | Black Truck is done!                   | rakesh wrote a section and finished the story! | do not                    | it's your turn   | do                   |
            | write and finish | gabe      | Black Truck is done!                   | rakesh wrote a section and finished the story! | do not                    | it's zach's turn | do                   |
            | finish           | zach      | Black Truck is done!                   | rakesh finished the story!                     | do not                    | it's your turn   | do not               |
            | finish           | gabe      | Black Truck is done!                   | rakesh finished the story!                     | do not                    | it's zach's turn | do not               |
            | pass             | zach      | It's your turn on Black Truck!         | rakesh passed!                                 | do                        | it's your turn   | do not               |
            | pass             | gabe      | rakesh took their turn on Black Truck! | rakesh passed!                                 | do                        | it's zach's turn | do not               |


