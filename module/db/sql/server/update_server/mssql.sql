UPDATE [server]
SET update_at = CURRENT_TIMESTAMP, [name] = @p2, software = @p3, [version] = @p4
WHERE id = @p1;
