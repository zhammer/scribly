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

    Scenario: Zach is notified that gabe added zach to a story
        Given I am logged in as gabe
        When I visit "/new"
        And I click on the "title" input
        And I type "The Horseman of Bonny Light"
        And I hit tab
        And I type "Oh Napolean Bonaparte, be the cause of my woe..."
        And I click the button "add cowriters"
        And I click on the "person-1" input
        And I type "zach"
        And I click the button "submit"
        And I wait .5 seconds
        And I log in as "zach"
        And I open my email at "zach@mail.com" with the subject "gabe started the story The Horseman of Bonny Light - it's your turn!"
        Then I see the text "gabe started the story The Horseman of Bonny Light!"
        And I see the text "it's your turn"
        And I see the text "Oh Napolean Bonaparte, be the cause of my woe..."
        When I click the link "The Horseman of Bonny Light"
        Then I am on "/stories/1"


