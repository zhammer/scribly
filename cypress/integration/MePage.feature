Feature: Me Page
    On the Me Page I can see my stories and I can start a new story.

    Background:
        Given the following users exist
            | username |
            | zach     |
            | gabe     |
            | rakesh   |

    Scenario Outline: I visit the Me Page as <username>
        Given I am logged in as <username>
        When I visit "/me"
        Then I see the text "<username>'s scribly"
        And I see the text "start a new story"

        Examples:
            | username |
            | zach     |
            | gabe     |


    Scenario: I try to visit the Me Page without being logged in
        Given I am not logged in
        When I visit "/me"
        Then I am on "/"
        And I see the text "log in"

    Scenario: I click start a new story
        Given I am logged in as zach
        When I visit "/me"
        And I click the text "start a new story"
        Then I am on "/new"
        And I see the text "add cowriters (saves current draft)"

    Scenario: I see my stories on the me page, organized by status
        Given the following stories exist
            | title             | turns | users        | complete |
            | The cool dog      | 8     | zach, gabe   | false    |
            | A big car         | 9     | zach, gabe   | true     |
            | Waiting for Dotty | 1     | zach         | false    |
            | Pushkins Theory   | 1     | rakesh, gabe | false    |
            | Debussys peanut   | 2     | gabe, zach   | false    |
        And I am logged in as zach
        When I visit "/me"
        Then the "drafts" section has the stories
            | title             |
            | Waiting for Dotty |
        And the "in progress" section has the stories
            | title           |
            | The cool dog    |
            | Debussys peanut |
        And the "done" section has the stories
            | title     |
            | A big car |

    Scenario Outline: I click on the story <storyTitle>
        Given the following stories exist
            | title             | turns | users      | complete |
            | The cool dog      | 8     | zach, gabe | false    |
            | A big car         | 9     | zach, gabe | true     |
            | Waiting for Dotty | 1     | zach       | false    |
        And I am logged in as zach
        When I visit "/me"
        And I click the text "<storyTitle>"
        Then I see the title "<storyTitle>"

        Examples:
            | storyTitle        |
            | The cool dog      |
            | A big car         |
            | Waiting for Dotty |
