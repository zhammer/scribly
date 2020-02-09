const argon2 = require("argon2");
const fs = require("fs");
const path = require("path");
const util = require("util");
const Database = require("better-sqlite3");
const casual = require("casual");

const DATABASE_URL = process.env_URL || ".data/scribly.db";
const DB_SCHEMA = fs.readFileSync("migrations/createdb.sql", "utf8");

const unlinkAsync = util.promisify(fs.unlink);

class DB {
  constructor() {
    this._client;
    this._passwordHash;

    // make sure dirname exists
    const dirname = path.dirname(DATABASE_URL);
    try {
      fs.mkdirSync(dirname, { recursive: true });
    } catch (e) {}
  }

  _getPasswordHash = async () => {
    if (this._passwordHash) return this._passwordHash;

    this._passwordHash = await argon2.hash("password");
    return this._passwordHash;
  };

  resetDb = async () => {
    try {
      await unlinkAsync(DATABASE_URL);
    } catch (e) {}
    const client = new Database(DATABASE_URL);
    client.exec(DB_SCHEMA);
    return null;
  };

  addStories = async stories => {
    for (const story of stories) {
      await this.addStory(story);
    }
    return null;
  };

  _fetchUsers = async usernames => {
    const client = new Database(DATABASE_URL);
    const stmt = client.prepare(
      "SELECT * FROM users WHERE username IN (" +
        usernames.map(username => `'${username}'`).join(", ") +
        ")"
    );
    const usersResult = stmt.all();
    return usernames.map(username =>
      usersResult.find(row => row.username === username)
    );
  };

  addStory = async storyInput => {
    const { title, turns, usernames, complete } = storyInput;
    const client = new Database(DATABASE_URL);
    const users = await this._fetchUsers(usernames);
    let state = "draft";
    if (users.length > 1) state = "in_progress";
    if (complete) state = "done";
    let stmt = client.prepare(
      `
        INSERT INTO stories (title, state, created_by)
        VALUES (@title, @state, @created_by);
    `
    );
    const { lastInsertRowid: storyId } = stmt.run({
      title,
      state,
      created_by: users[0].id
    });

    if (usernames.length > 1) {
      let stmt = client.prepare(`
              INSERT INTO story_cowriters (story_id, user_id, turn_index)
              VALUES (@storyId, @userId, @index)
        `);
      // add cowriters
      for (const [index, user] of users.entries()) {
        stmt.run({ storyId, userId: user.id, index });
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

    stmt = client.prepare(`
        INSERT INTO turns (story_id, taken_by, action, text_written)
        VALUES (?, ?, ?, ?)
    `);
    for (const turn of generatedTurns) {
      stmt.run([storyId, turn.user.id, turn.action, turn.text]);
    }

    return null;
  };

  addUsers = async users => {
    const passwordHash = await this._getPasswordHash();
    const client = new Database(DATABASE_URL);
    const insert = client.prepare(
      "INSERT INTO users (username, password, email, email_verification_status) VALUES (@username, @password, @email, @email_verification_status)"
    );
    users.forEach(user => {
      insert.run({
        ...user,
        email: `${user.username}@mail.com`,
        password: passwordHash
      });
    });
    return null;
  };
}

function odds(likelihood) {
  return Math.random() <= likelihood;
}

function range(length) {
  return [...Array(length).keys()];
}

module.exports = DB;
