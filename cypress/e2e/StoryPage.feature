Feature: Story Page
    # Most story functionality is covered in StoryTurn.feature

    Scenario Outline: I need an accessible story page when <condition>
        Given the following users exist
            | username |
            | zach     |
            | gabe     |
        And the following stories exist
            | title                        | turns   | users      | complete   |
            | higgs on the petal of a leaf | <turns> | zach, gabe | <complete> |
        And I am logged in as zach
        When I visit "/stories/1"
        Then the page is accessible

        Examples:
            | condition               | turns | complete |
            | it is my turn           | 2     | false    |
            | it is my cowriters turn | 1     | false    |
            | the story is finished   | 2     | true     |
