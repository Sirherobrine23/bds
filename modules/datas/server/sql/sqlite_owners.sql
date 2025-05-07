SELECT p.user_id, p.permission, u.name, u.username, u.permission
FROM servers_permission AS p
JOIN user AS u ON u.id = p.user_id
WHERE server_id = $1