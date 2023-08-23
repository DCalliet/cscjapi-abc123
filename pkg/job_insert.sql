INSERT INTO "job" (
  data,
  status,
  created_at
) VALUES (?, ?, ?) RETURNING "job"."id"