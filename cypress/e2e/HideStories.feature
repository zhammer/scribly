Feature: Hide Stories
    I can hide stories that I'm not interested in seeing on my Me Page.
    I can view stoires I've hidden and unhide them at a later point.

    Background:
        Given the following users exist
            | username |
            | zach     |
            | gabe     |

        And the following stories exist
            | title               | turns | users      | complete |
            | Linger on           | 3     | zach, gabe | false    |
            | Your pale blue eyes | 3     | zach, gabe | false    |

    Scenario: I hide a story and unhide a story
        Given I am logged in as zach
        When I visit "/me"
        And I click the hide button for the story "Linger on"
        Then I do not see the text "Linger on"

        When I click the link "show hidden stories"
        Then I see the text "(hidden) Linger on"

        When I click the unhide button for the story "Linger on"
        And I click the link "don't show hidden stories"
        Then I see the text "Linger on"

    Scenario: I hide multiple stories
        Given I am logged in as zach
        When I visit "/me"
        And I click the hide button for the story "Linger on"
        And I click the hide button for the story "Your pale blue eyes"
        Then I do not see the text "Linger on"
        And I do not see the text "Your pale blue eyes"