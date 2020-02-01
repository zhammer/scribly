Feature: Nudge
    I can nudge the current writer on a story to take their turn.

    Scenario Outline: I nudge the current writer, <currentWriter>
        Given the following users exist
            | username | email_verification_status |
            | zach     | verified                  |
            | gabe     | verified                  |
            | rakesh   | verified                  |
        And the following stories exist
            | title       | turns   | users              | complete |
            | Assume Form | <turns> | zach, gabe, rakesh | false    |
        And I am logged in as zach
        When I visit "/stories/1"
        And I click the button "nudge <currentWriter> to take their turn"
        Then I see the text "zach, your nudge has been delivered."

        When I click the link "go back to story"
        Then I am on "/stories/1"

        When I log in as "<currentWriter>"
        And I open my email at "<currentWriter>@mail.com" with the subject "zach nudged you to take your turn on Assume Form"
        And I click the link "Assume Form"
        Then I am on "/stories/1"

        Examples:
            | turns | currentWriter |
            | 1     | gabe          |
            | 2     | rakesh        |

    Scenario Outline: I don't see the nudge button when <scenario>
        Given the following users exist
            | username | email_verification_status |
            | zach     | verified                  |
            | gabe     | verified                  |
        And the following stories exist
            | title       | turns | users      | complete   |
            | Assume Form | 2     | zach, gabe | <complete> |
        And I am logged in as zach
        When I visit "/stories/1"
        Then I do not see the text "nudge"

        Examples:
            | scenario              | complete |
            | it's my turn          | false    |
            | the story is finished | true     |

    Scenario Outline: Nudge fails
