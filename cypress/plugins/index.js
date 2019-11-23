const argon2 = require("argon2");
const fs = require("fs");
const cucumber = require("cypress-cucumber-preprocessor").default;
const { Client } = require("pg");

const DATABASE_URL = process.env.DATABASE_URL || "postgres://localhost/scribly";
const DB_SCHEMA = fs.readFileSync("migrations/createdb.sql", "utf8");

let _client;
async function getClient() {
  if (_client) return _client;

  _client = new Client(DATABASE_URL);
  await _client.connect();
  return _client;
}

let _hashedPassword;
async function getHashedPassword() {
  if (_hashedPassword) return _hashedPassword;
  _hashedPassword = await argon2.hash("password");
  return _hashedPassword;
}

async function resetDb() {
  const client = await getClient();
  return await client.query(`
    DROP SCHEMA IF EXISTS public CASCADE;
    CREATE SCHEMA public;
    ${DB_SCHEMA}
  `);
}

async function addUsers(users) {
  let hashedPassword = await getHashedPassword();
  const values = users.reduce(
    ([usernames, passwords, emails], { username, email }) => [
      [...usernames, username],
      [...passwords, hashedPassword],
      [...emails, email]
    ],
    [[], [], []]
  );

  const client = await getClient();
  return await client.query(
    `
        INSERT INTO users (username, password, email)
        SELECT * FROM UNNEST ($1::text[], $2::text[], $3::text[])
    `,
    values
  );
}

module.exports = on => {
  on("file:preprocessor", cucumber());
  on("task", {
    resetDb,
    addUsers
  });
};
