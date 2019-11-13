Cypress.Commands.add("resetdb", () => {
  cy.exec("pipenv run python scripts/createdb.py --reset", {
    failOnNonZeroExit: false
  }).then(result => {
    console.log(result.stdout);
    console.log(result.stderr);
    if (result.code) {
      throw result.stderr;
    }
  });
});

Cypress.Commands.add("addusers", users => {
  const usersArg = users
    .map(({ username, password }) => `${username}:${password}`)
    .join(" ");
  cy.exec(`pipenv run python scripts/addusers.py ${usersArg}`);
});
