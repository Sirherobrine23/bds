INSERT INTO user(username, name, email)
VALUES (lower($1), $2, lower($3));
