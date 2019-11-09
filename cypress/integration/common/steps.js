/* global cy */
/// <reference types="cypress" />
import { Given, Then, When } from "cypress-cucumber-preprocessor/steps";

When(`I visit {string}`, path => {
  cy.visit(path);
});

Then(`I see the text {string}`, text => {
  cy.contains(text);
});
