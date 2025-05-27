DELETE FROM [friends]
WHERE server_id = @p1 AND user_id = @p2;
