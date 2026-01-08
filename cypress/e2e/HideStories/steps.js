/* global cy */
/// <reference types="cypress" />
import { Given, Then, When } from "@badeball/cypress-cucumber-preprocessor";

When(
  /I click the (hide|unhide) button for the story "(.*)"/,
  (hideAction, storyTitle) => {
    cy.get("li")
      .contains(storyTitle)
      .get("button")
      .contains(hideAction)
      .click();
  }
);
