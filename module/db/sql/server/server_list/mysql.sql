SELECT id,
  `name`,
  `owner`,
  software,
  `version`,
  create_at,
  update_at
FROM `server`
WHERE `server`.`owner` = ?
  OR `server`.`id` IN (
    SELECT f.server_id
    FROM `friends` f
    WHERE f.`user_id` = ?
      AND JSON_CONTAINS(f.`permissions`, '"view"')
  );
