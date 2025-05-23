SELECT id,
  name,
  owner,
  software,
  version,
  create_at,
  update_at
FROM server
WHERE server.owner = $1
  OR server.id IN (
    SELECT server_id
    FROM friends,
      json_each(friends.permissions)
    WHERE user_id = $1
      AND json_each.value = 'view'
  );