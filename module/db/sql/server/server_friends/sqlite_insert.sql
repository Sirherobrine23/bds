-- Insert in the table if server_id does not have user_id, if you ignore, user_id should be unique for each server
INSERT INTO friends(server_id, user_id, permissions)
VALUES ($1, $2, $3);