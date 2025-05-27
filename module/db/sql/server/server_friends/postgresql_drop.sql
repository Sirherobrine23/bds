DELETE FROM "friends"
WHERE server_id = $1 AND user_id = $2;
