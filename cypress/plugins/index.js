const cucumber = require("cypress-cucumber-preprocessor").default;
const DB = require("./db");

module.exports = on => {
  const db = new DB();
  on("file:preprocessor", cucumber());
  on("task", {
    resetDb: db.resetDb,
    addUsers: db.addUsers
  });
};
