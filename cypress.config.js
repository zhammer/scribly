const { defineConfig } = require("cypress");
const createBundler = require("@bahmutov/cypress-esbuild-preprocessor");
const {
  addCucumberPreprocessorPlugin,
} = require("@badeball/cypress-cucumber-preprocessor");
const {
  createEsbuildPlugin,
} = require("@badeball/cypress-cucumber-preprocessor/esbuild");
const DB = require("./cypress/plugins/db");
const MockResend = require("./cypress/plugins/mockresend");

module.exports = defineConfig({
  projectId: "nxmvpu",
  e2e: {
    baseUrl: "http://127.0.0.1:8000",
    specPattern: "cypress/e2e/**/*.feature",
    async setupNodeEvents(on, config) {
      // Cucumber preprocessor
      await addCucumberPreprocessorPlugin(on, config);

      on(
        "file:preprocessor",
        createBundler({
          plugins: [createEsbuildPlugin(config)],
        })
      );

      // Custom tasks from old plugins/index.js
      const db = new DB();
      const resend = new MockResend(9991);

      on("task", {
        resetDb: db.resetDb,
        addUsers: db.addUsers,
        addStories: db.addStories,
        listenForEmails: resend.listenForEmails,
        getEmails: resend.getEmails,
      });

      return config;
    },
  },
});
