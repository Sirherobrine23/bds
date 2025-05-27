UPDATE `server`
SET update_at = CURRENT_TIMESTAMP, `name` = ?, software = ?, `version` = ?
WHERE id = ?;
