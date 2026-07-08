-- +goose Up

-- Tighten the projects and casefiles tenant scope now that their writers always
-- supply a tenant_id: projects are minted under the caller's tenant (server: the
-- authenticated principal; CLI: the fixed local tenant sentinel) and casefiles
-- inherit their tenant as part of aggregate identity. Both repositories persist
-- tenant_id from the aggregate, so the column can no longer be null. Pleadings
-- stay nullable until their own slice wires the tenant end to end.

ALTER TABLE projects ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE casefiles ALTER COLUMN tenant_id SET NOT NULL;
