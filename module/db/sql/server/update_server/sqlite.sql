UPDATE server
SET update_at = current_timestamp, name = $2, software = $3, version = $4
WHERE id = $1;