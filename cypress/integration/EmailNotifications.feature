Feature: Email Notifications
    Scribly users get email notifications to keep them updated on their
    stories.

    Background:
        Given the following users exist
            | username | email_verification_status |
            | zach     | verified                  |
            | gabe     | verified                  |
            | rakesh   | verified                  |

    Scenario: Email notifications are sent after someone takes a turn
        Given the following stories exist
            | title       | turns | users              | complete |
            | Black Truck | 5     | zach, gabe, rakesh | false    |
        And I am logged in as rakesh
        When I visit "/stories/1"
        And I click on the "text" textarea
        And I type "I've been through the fires. I've felt embers down my spine."
        And I click the button "write"
        And I wait 1 seconds
        And I open my email at "zach@mail.com" with the subject "It's your turn on Black Truck!"
