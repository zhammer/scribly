/* global cy */
/// <reference types="cypress" />
import { Given, Then, When } from "cypress-cucumber-preprocessor/steps";

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
