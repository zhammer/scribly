Cypress.Commands.add("resetDb", () => {
  cy.task("resetDb");
});

Cypress.Commands.add("addUsers", users => {
  cy.task("addUsers", users);
});

Cypress.Commands.add("addStories", stories => {
  cy.task("addStories", stories);
});

Cypress.Commands.add("listenForEmails", () => {
  cy.task("listenForEmails");
});
