-- Main table
CREATE TABLE IF NOT EXISTS "user" (
  id SERIAL PRIMARY KEY,
  name text NOT NULL,
  username varchar(25) UNIQUE NOT NULL,
  permission INTEGER DEFAULT 0
);

-- Table to storage passwords
CREATE TABLE IF NOT EXISTS "password" (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES public.user(id),
  password text
);

-- Table to storage user tokens
CREATE TABLE IF NOT EXISTS "token" (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES public.user(id),
  token text NOT NULL,
  permission INTEGER DEFAULT 0
);
