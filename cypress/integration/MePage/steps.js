/* global cy */
/// <reference types="cypress" />
import { Given, Then, When } from "cypress-cucumber-preprocessor/steps";

Then(`the {string} section has the stories`, (sectionTitle, datatable) => {
  const stories = datatable.hashes().map(row => row.title);
  cy.get("section")
    .contains("h2", sectionTitle)
    .parent()
    .within(() => {
      stories.forEach(storyTitle => {
        cy.get("li").contains(storyTitle);
      });
    });
  cy.log(sectionTitle);
});
