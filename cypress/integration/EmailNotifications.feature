Feature: Email Notifications
    Scribly users get email notifications to keep them updated on their
    stories.

    Background:
        Given the following users exist
            | username | email_verification_status |
            | zach     | verified                  |
            | gabe     | verified                  |
            | rakesh   | verified                  |

    Scenario Outline: Email notifications are sent after someone takes a turn
        Given the following stories exist
            | title       | turns | users              | complete |
            | Black Truck | 5     | zach, gabe, rakesh | false    |
        And I am logged in as rakesh
        When I visit "/stories/1"
        And I click on the "text" textarea
        And I type "I've been through the fires. I've felt embers down my spine."
        And I click the button "write"
        And I log in as "<recipient>"
        And I open my email at "<recipient>@mail.com" with the subject "<subject>"
        Then I see the text "rakesh wrote a section!"
        And I see the text "<emailText>"
        And I see the text "I've been through the fires. I've felt embers down my spine."

        Examples:
            | recipient | subject                                | emailText1              | emailText        |
            | zach      | It's your turn on Black Truck!         | rakesh wrote a section! | it's your turn   |
            | gabe      | rakesh took their turn on Black Truck! | rakesh wrote a section! | it's zach's turn |

