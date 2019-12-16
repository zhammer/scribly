const cucumber = require("cypress-cucumber-preprocessor").default;
const DB = require("./db");
const MockSendGrid = require("./mocksendgrid");

module.exports = on => {
  const db = new DB();
  const sendgrid = new MockSendGrid(9991);
  on("file:preprocessor", cucumber());
  on("task", {
    resetDb: db.resetDb,
    addUsers: db.addUsers,
    addStories: db.addStories,
    listenForEmails: sendgrid.listenForEmails,
    getEmails: sendgrid.getEmails
  });
};
