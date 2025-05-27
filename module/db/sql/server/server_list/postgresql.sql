SELECT id,
  "name",
  "owner",
  software,
  "version",
  create_at,
  update_at
FROM "server"
WHERE "server"."owner" = $1
  OR "server"."id" IN (
    SELECT f.server_id
    FROM "friends" f
    WHERE f."user_id" = $1
      AND f."permissions" @> '"view"'::jsonb
  );
