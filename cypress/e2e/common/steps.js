/* global cy */
/// <reference types="cypress" />
import { Given, Then, When } from "@badeball/cypress-cucumber-preprocessor";

beforeEach(() => {
  cy.resetDb();
  cy.listenForEmails();
});

Given("the following users exist", datatable => {
  const users = datatable.hashes().map(row => ({
    username: row.username,
    email_verification_status: row.email_verification_status || "verified"
  }));
  cy.addUsers(users);
});

Given("I am not logged in", () => { });

Given(/I am logged in as (.*)/, username => {
  cy.clearCookies();
  cy.visit("/login");
  cy.get("input[name='username']").type(username);
  cy.get("input[name='password']").type("password");
  cy.get("button").click();
  cy.location("pathname").should("eq", "/me");
});

Given("the following stories exist", datatable => {
  cy.addStories(
    datatable.hashes().map(storyRow => ({
      ...storyRow,
      usernames: storyRow.users.split(", "),
      complete: storyRow.complete === "true" ? true : false,
      turns: parseInt(storyRow.turns)
    }))
  );
});

When("I wait {float} seconds", seconds => {
  cy.wait(seconds * 1000);
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

When(`I visit {string} expecting a non-200 response`, path => {
  cy.visit(path, { failOnStatusCode: false });
});

When(/I click the (text|button|link) "(.*)"/, (elementType, text) => {
  const regex = new RegExp(text);
  console.log(regex);
  const mapping = {
    link: "a",
    button: "button"
  };
  if (elementType === "text") {
    cy.contains(regex).click();
  } else {
    cy.get(mapping[elementType])
      .contains(regex)
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
  cy.clearCookies(); // cy.visit("/logout");

  cy.visit("/login");
  cy.get("input[name='username']").type(username);
  cy.get("input[name='password']").type("password");
  cy.get("button").click();
  cy.location("pathname").should("eq", "/me");
});

Then(/I (?:do )?see the text "(.*)"/, text => {
  cy.contains(text);
});

Then(`I do not see the text {string}`, text => {
  cy.get("body")
    .contains(text)
    .should("not.exist");
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

Then(`I see the title {string}`, title => {
  cy.get("h1").contains(title);
});

function getEmail(emails, expectedAddress, expectedSubject) {
  return emails.find(email => {
    return email.personalizations.find(({ to, subject }) => {
      return (
        subject === expectedSubject &&
        to.find(({ email }) => email === expectedAddress)
      );
    });
  });
}

Then(
  /(?:I received|there is) an email at "(.*)" with the subject "(.*)"/,
  (address, subject) => {
    cy.getEmails().then(emails => {
      const email = getEmail(emails, address, subject);
      expect(email).to.exist;
    });
  }
);

When(
  "I open my email at {string} with the subject {string}",
  (address, subject) => {
    cy.getEmails().then(emails => {
      const email = getEmail(emails, address, subject);
      const html = email.content[0].value;
      cy.writeFile("static/tempfile.html", html);
      cy.visit("_cypress_email");
    });
  }
);

Then("the page is accessible", () => {
  cy.injectAxe();
  cy.checkA11y();
});

Then(`the text {string} is in the viewport`, text => {
  cy.contains(text).should("be.inViewport");
});
