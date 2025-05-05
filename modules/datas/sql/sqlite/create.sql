PRAGMA foreign_keys = ON;

-- Main table
CREATE TABLE IF NOT EXISTS "user" (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name text NOT NULL,
  username varchar(25) UNIQUE NOT NULL,
  permission INTEGER DEFAULT 0
);

-- Table to storage passwords
CREATE TABLE IF NOT EXISTS "password" (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER UNIQUE REFERENCES user (id) ON DELETE CASCADE,
  password text
);

-- Table to storage user tokens
CREATE TABLE IF NOT EXISTS "token" (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER REFERENCES user (id) ON DELETE CASCADE,
  token text NOT NULL UNIQUE,
  permission INTEGER DEFAULT 0
);
