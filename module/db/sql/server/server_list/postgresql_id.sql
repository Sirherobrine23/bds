SELECT id, "name", "owner", software, "version", create_at, update_at
FROM "server"
WHERE id = $1;
