import "./commands";
import "./assertions";
import "cypress-axe";
import "cypress-real-events";

// Add tab command for compatibility with old cypress-plugin-tab
Cypress.Commands.add("tab", { prevSubject: "element" }, (subject) => {
  cy.wrap(subject).realPress("Tab");
  return cy.focused();
});
