-- +goose Up

-- Tighten the pleadings tenant scope now that its writer always supplies a
-- tenant_id: pleadings carry the tenant as part of aggregate identity (filed
-- under the caller's tenant — server: the authenticated principal), and the
-- repository persists pleadings.tenant_id from the aggregate. This is the last
-- judicial aggregate to be scoped, so every judicial table is now NOT NULL.

ALTER TABLE pleadings ALTER COLUMN tenant_id SET NOT NULL;
