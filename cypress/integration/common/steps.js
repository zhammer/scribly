/* global cy */
/// <reference types="cypress" />
import { Given, Then, When } from "cypress-cucumber-preprocessor/steps";

beforeEach(() => {
  cy.resetdb();
});

Given("the following users exist", datatable => {
  const users = datatable.hashes().map(row => ({
    username: row.username,
    password: "password",
    email: `${row.username}@mail.com`
  }));
  cy.addusers(users);
});

Given("I am not logged in", () => {});

Given(/I am logged in as (.*)/, username => {
  cy.visit("/login");
  cy.get("input[name='username']").type(username);
  cy.get("input[name='password']").type("password");
  cy.get("button").click();
  cy.location("pathname").should("eq", "/me");
});

When("I hit tab", () => {
  cy.focused().tab();
});

When("I refresh the page", () => {
  cy.reload();
});

When(`I visit {string}`, path => {
  cy.visit(path);
});

When(/I click the (text|button) "(.*)"/, (elementType, text) => {
  if (elementType === "text") {
    cy.contains(text).click();
  } else {
    cy.get(elementType)
      .contains(text)
      .click();
  }
});

When(/I click on the "(.*)" (input|textarea)/, (name, formElement) => {
  cy.get(`${formElement}[name='${name}']`).click();
});

When(`I type {string}`, text => {
  cy.focused().type(text);
});

When("I type:", text => {
  cy.focused().type(text);
});

When(`I log in as {string}`, username => {
  cy.visit("/logout");

  cy.visit("/login");
  cy.get("input[name='username']").type(username);
  cy.get("input[name='password']").type("password");
  cy.get("button").click();
  cy.location("pathname").should("eq", "/me");
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

Then(/I (can|cannot) see the turn form/, canOrCannot => {
  const should = canOrCannot === "can" ? "exist" : "not.exist";
  cy.get("#turn-form").should(should);
});
