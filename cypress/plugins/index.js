const argon2 = require("argon2");
const fs = require("fs");
const cucumber = require("cypress-cucumber-preprocessor").default;
const { Client } = require("pg");

let shared_client;
async function getClient(dbUrl) {
  if (shared_client) return shared_client;

  shared_client = new Client(dbUrl);
  await shared_client.connect();
  return shared_client;
}

async function resetDb(dbUrl, dbSchema) {
  const client = await getClient(dbUrl);
  return await client.query(`
    DROP SCHEMA IF EXISTS public CASCADE;
    CREATE SCHEMA public;
    ${dbSchema}
  `);
}

async function addUsers(dbUrl, users) {
  let hashedPassword = await argon2.hash("password");
  const values = users.reduce(
    ([usernames, passwords, emails], { username, email }) => [
      [...usernames, username],
      [...passwords, hashedPassword],
      [...emails, email]
    ],
    [[], [], []]
  );

  const client = await getClient(dbUrl);
  return await client.query(
    `
        INSERT INTO users (username, password, email)
        SELECT * FROM UNNEST ($1::text[], $2::text[], $3::text[])
    `,
    values
  );
}

const DB_SCHEMA = fs.readFileSync("migrations/createdb.sql", "utf8");
module.exports = (on, config) => {
  on("file:preprocessor", cucumber());
  on("task", {
    resetDb() {
      return resetDb(config.env.DATABASE_URL, DB_SCHEMA);
    },
    addUsers(users) {
      return addUsers(config.env.DATABASE_URL, users);
    }
  });
};
