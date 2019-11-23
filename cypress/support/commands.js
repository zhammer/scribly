Cypress.Commands.add("resetdb", () => {
  cy.task("resetDb");
});

Cypress.Commands.add("addusers", users => {
  cy.task("addUsers", users);
});
