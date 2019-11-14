Feature: Story Turn
    Users write stories together by taking turns. The types of turns are:
    - write: write a section and add it to the story.
    - pass: pass on your turn without writing any text.
    - finish: finish the story without writing any text.
    - write_and_finish: write a section and add it to the story, and finish the story.

    Background:
        Given the following users exist
            | username |
            | zach     |
            | gabe     |
            | rakesh   |
        And I am logged in as zach
        And I visit "/new"
        And I click on the "title" input
        And I type "A warm box of pizza, a rowboat"
        And I hit tab
        And I type:
            """
            Italy is beautiful in the summers. The warm nights. The stars.
            The box of pizza Alvinci and Moricori cracked open, just a smidge,
            to smell that sweet pizza smell on the water...
            """
        And I click the text "add cowriters"
        And I click on the "person-1" input
        And I type "gabe"
        And I hit tab
        And I type "rakesh"
        And I click the text "submit"

    Scenario Outline: On turn 1, only gabe (whose turn it is) can see text input
        Given I am logged in as <username>
        When I visit "/stories/1"
        Then I see the text "A warm box of pizza, a rowboat"
        And I see the text "Italy is beautiful in the summers."
        And I see the text "cowriters: zach, gabe, rakesh"
        And I see the text "turn: 1 (gabe's turn)"
        And I <canOrCannot> see the turn form

        Examples:
            | username | canOrCannot |
            | zach     | cannot      |
            | gabe     | can         |
            | rakesh   | cannot      |
