import "./commands";
import "./assertions";
import "cypress-axe";

// Add tab command for compatibility with old cypress-plugin-tab
// Uses Cypress's native type command with special {tab} sequence (requires experimental flag)
// Falls back to finding next focusable element if that doesn't work
Cypress.Commands.add("tab", { prevSubject: "element" }, (subject) => {
  // Get all focusable elements on the page
  const focusableSelector = 'input:not([disabled]), textarea:not([disabled]), button:not([disabled]), select:not([disabled]), a[href], [tabindex]:not([tabindex="-1"])';

  return cy.get(focusableSelector).then($focusable => {
    const focusableArray = $focusable.toArray();
    const currentIndex = focusableArray.indexOf(subject[0]);
    const nextIndex = currentIndex + 1;

    if (nextIndex < focusableArray.length) {
      return cy.wrap(focusableArray[nextIndex]).focus();
    }
    return cy.wrap(subject);
  });
});
