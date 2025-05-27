SELECT id, server_id, user_id, permissions
FROM [friends]
WHERE server_id = @p1;
