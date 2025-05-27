SELECT id, server_id, uuid, software, version, create_at
FROM "backups"
WHERE server_id = $1;
