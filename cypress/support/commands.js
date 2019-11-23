Cypress.Commands.add("resetdb", () => {
  cy.task("resetDb");
});

Cypress.Commands.add("addUsers", users => {
  cy.task("addUsers", users);
});
