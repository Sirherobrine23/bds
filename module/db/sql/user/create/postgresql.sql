INSERT INTO "user" (username, "name", email)
VALUES (LOWER($1), $2, LOWER($3));
