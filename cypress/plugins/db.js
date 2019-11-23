const argon2 = require("argon2");
const fs = require("fs");
const { Client } = require("pg");

const DATABASE_URL = process.env.DATABASE_URL || "postgres://localhost/scribly";
const DB_SCHEMA = fs.readFileSync("migrations/createdb.sql", "utf8");

class DB {
  constructor() {
    this._client;
    this._passwordHash;
  }

  _getClient = async () => {
    if (this._client) return this._client;

    this._client = new Client(DATABASE_URL);
    await this._client.connect();
    return this._client;
  };

  _getPasswordHash = async () => {
    if (this._passwordHash) return this._passwordHash;

    this._passwordHash = await argon2.hash("password");
    return this._passwordHash;
  };

  resetDb = async () => {
    const client = await this._getClient();
    return await client.query(`
      DROP SCHEMA IF EXISTS public CASCADE;
      CREATE SCHEMA public;
      ${DB_SCHEMA}
    `);
  };

  addUsers = async usernames => {
    const passwordHash = await this._getPasswordHash();
    const nestedRows = usernames.reduce(
      ([usernames, passwords, emails], username) => [
        [...usernames, username],
        [...passwords, passwordHash],
        [...emails, `${username}@mail.com`]
      ],
      [[], [], []]
    );

    const client = await this._getClient();
    return await client.query(
      `
            INSERT INTO users (username, password, email)
            SELECT * FROM UNNEST ($1::text[], $2::text[], $3::text[])
        `,
      nestedRows
    );
  };
}

module.exports = DB;
