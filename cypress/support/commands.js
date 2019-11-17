Cypress.Commands.add("resetdb", () => {
  cy.exec("pipenv run python scripts/createdb.py --reset");
});

Cypress.Commands.add("addusers", users => {
  const usersArg = users
    .map(({ username, password, email }) => `${username}:${password}:${email}`)
    .join(" ");
  cy.exec(`pipenv run python scripts/addusers.py ${usersArg}`);
});
