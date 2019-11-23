const argon2 = require("argon2");
const fs = require("fs");
const { Client } = require("pg");
const casual = require("casual");

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

  addStories = async stories => {
    for (const story of stories) {
      await this.addStory(story);
    }
    return null;
  };

  _fetchUsers = async usernames => {
    const client = await this._getClient();
    const usersResult = await client.query(
      `
          SELECT * FROM users
          WHERE username = ANY($1::text[])

      `,
      [usernames]
    );
    return usernames.map(username =>
      usersResult.rows.find(row => row.username === username)
    );
  };

  addStory = async storyInput => {
    const { title, turns, usernames, complete } = storyInput;
    const client = await this._getClient();
    const users = await this._fetchUsers(usernames);
    let state = "draft";
    if (users.length > 1) state = "in_progress";
    if (complete) state = "done";
    const story = await client.query(
      `
        INSERT INTO stories (title, state, created_by)
        VALUES ($1, $2, $3)
        RETURNING *
    `,
      [title, state, users[0].id]
    );
    const storyId = story.rows[0].id;

    if (usernames.length > 1) {
      // add cowriters
      for (const [index, user] of users.entries()) {
        await client.query(
          `
                INSERT INTO story_cowriters (story_id, user_id, turn_index)
                VALUES ($1, $2, $3)
            `,
          [storyId, user.id, index]
        );
      }
    }

    const generatedTurns = range(turns).map(turnIndex => {
      const user = users[turnIndex % users.length];
      // first turn always a write
      if (turnIndex === 0) {
        return {
          user,
          action: "write",
          text: casual.text
        };
      }

      // if last turn and story complete, a write or write and finish
      if (complete && turnIndex === turns - 1) {
        const action = odds(0.5) ? "write_and_finish" : "finish";
        return {
          user,
          action,
          text: action === "finish" ? "" : casual.text
        };
      }

      // otherwise it's either a write or a pass
      const action = odds(0.1) ? "pass" : "write";
      return {
        user,
        action,
        text: action === "pass" ? "" : casual.text
      };
    });

    for (const turn of generatedTurns) {
      await client.query(
        `
            INSERT INTO turns (story_id, taken_by, action, text_written)
            VALUES ($1, $2, $3, $4)
        `,
        [storyId, turn.user.id, turn.action, turn.text]
      );
    }

    return null;
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

function odds(likelihood) {
  return Math.random() <= likelihood;
}

function range(length) {
  return [...Array(5).keys()];
}

module.exports = DB;
