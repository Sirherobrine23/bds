-- User
INSERT INTO user(username, name, email)
VALUE (LOWER($1), $2, LOWER($3));
-- Password hash
INSERT INTO password(user, password)
VALUE ((SELECT id FROM user WHERE username = $1 LIMIT 1), $4);