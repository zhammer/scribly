Cypress.Commands.add("resetdb", () => {
  cy.exec("pipenv run python scripts/createdb.py --reset");
});

Cypress.Commands.add("addusers", users => {
  const usersArg = users
    .map(({ username, password }) => `${username}:${password}`)
    .join(" ");
  cy.exec(`pipenv run python scripts/addusers.py ${usersArg}`);
});
