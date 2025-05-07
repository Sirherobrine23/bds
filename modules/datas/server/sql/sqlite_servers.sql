SELECT servers.id,
  servers.name,
  servers.server_type,
  servers.server_version
FROM servers_permission AS permission
  JOIN servers AS servers ON servers.id = permission.server_id
WHERE user_id = $1 -- user id
  AND permission = $2 -- Owner permission