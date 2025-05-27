SELECT id,
  [name],
  [owner],
  software,
  [version],
  create_at,
  update_at
FROM [server]
WHERE [server].[owner] = @p1
  OR [server].[id] IN (
    SELECT f.server_id
    FROM [friends] f
    CROSS APPLY OPENJSON(f.[permissions]) AS p
    WHERE f.[user_id] = @p1
      AND p.[value] = 'view'
  );
