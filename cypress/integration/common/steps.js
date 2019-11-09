/* global cy */
/// <reference types="cypress" />
import { Given, Then, When } from "cypress-cucumber-preprocessor/steps";

beforeEach(() => {
  cy.wrap(false).as("loggedIn");
});

Given(/I (am|am not) logged in/, amOrAmNot => {
  const loggedIn = amOrAmNot === "am";
  cy.wrap(loggedIn).as("loggedIn");
});

When(`I visit {string}`, path => {
  cy.get("@loggedIn").then(loggedIn => {
    if (loggedIn) {
      cy.visit(path, {
        auth: { username: "zach.the.hammer@gmail.com", password: "password" }
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
