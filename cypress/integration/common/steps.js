/* global cy */
/// <reference types="cypress" />
import { Given, Then, When } from "cypress-cucumber-preprocessor/steps";

beforeEach(() => {
  cy.resetdb();
  cy.wrap(null).as("loggedInUser");
});

Given("the following users exist", datatable => {
  const users = datatable.hashes().map(row => ({
    username: row.username,
    password: row.password
  }));
  cy.addusers(users);
});

Given("I am not logged in", () => {});

Given(/I am logged in as (.*):(.*)/, (username, password) => {
  cy.wrap({ username, password }).as("loggedInUser");
});

When(`I visit {string}`, path => {
  cy.get("@loggedInUser").then(loggedInUser => {
    if (loggedInUser) {
      cy.visit(path, {
        auth: loggedInUser
      });
    } else {
      cy.visit(path);
    }
  });
});

When(`I click the text {string}`, text => {
  cy.contains(text).click();
});

When(/I click on the "(.*)" (input|textarea)/, (name, formElement) => {
  cy.get(`${formElement}[name='${name}']`).click();
});

When(`I type {string}`, text => {
  cy.focused().type(text);
});

Then(`I see the text {string}`, text => {
  cy.contains(text);
});

Then(`I see the button {string}`, text => {
  cy.get("button").contains(text);
});

Then(`I am on {string}`, path => {
  cy.location("pathname").should("eq", path);
});
