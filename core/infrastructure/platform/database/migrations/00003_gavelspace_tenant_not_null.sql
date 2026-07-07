-- +goose Up

-- Tighten the gavelspaces tenant scope now that its writer always supplies a
-- tenant_id: the create use case mints the aggregate under the caller's tenant
-- (server: the authenticated principal; CLI: the fixed local tenant sentinel),
-- and the repository persists gavelspaces.tenant_id from the aggregate. The
-- remaining judicial tables (projects, casefiles, pleadings) stay nullable until
-- their own slice wires the tenant end to end.

ALTER TABLE gavelspaces ALTER COLUMN tenant_id SET NOT NULL;
