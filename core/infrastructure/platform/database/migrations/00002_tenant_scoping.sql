-- +goose Up

-- Multi-tenancy: scope the judicial aggregates by tenant. Each aggregate root
-- table (gavelspaces, projects, casefiles, pleadings) gains a tenant_id
-- referencing iam_tenants. The column is added nullable so the existing
-- repositories keep inserting while each aggregate's slice is wired to supply
-- and require it; a later step tightens the column to NOT NULL once its writers
-- always populate it.
--
-- Existing rows are backfilled to the default tenant, matched by slug because
-- its id is a per-install random UUID (minted in-domain at first boot), not a
-- fixed constant. On a fresh database there are no judicial rows and no default
-- tenant yet, so the UPDATEs are no-ops.

ALTER TABLE gavelspaces ADD COLUMN tenant_id UUID;
ALTER TABLE projects ADD COLUMN tenant_id UUID;
ALTER TABLE casefiles ADD COLUMN tenant_id UUID;
ALTER TABLE pleadings ADD COLUMN tenant_id UUID;

UPDATE gavelspaces SET tenant_id = (SELECT id FROM iam_tenants WHERE slug = 'default') WHERE tenant_id IS NULL;
UPDATE projects SET tenant_id = (SELECT id FROM iam_tenants WHERE slug = 'default') WHERE tenant_id IS NULL;
UPDATE casefiles SET tenant_id = (SELECT id FROM iam_tenants WHERE slug = 'default') WHERE tenant_id IS NULL;
UPDATE pleadings SET tenant_id = (SELECT id FROM iam_tenants WHERE slug = 'default') WHERE tenant_id IS NULL;

ALTER TABLE gavelspaces
    ADD CONSTRAINT gavelspaces_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES iam_tenants(id) ON DELETE CASCADE;
ALTER TABLE projects
    ADD CONSTRAINT projects_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES iam_tenants(id) ON DELETE CASCADE;
ALTER TABLE casefiles
    ADD CONSTRAINT casefiles_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES iam_tenants(id) ON DELETE CASCADE;
ALTER TABLE pleadings
    ADD CONSTRAINT pleadings_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES iam_tenants(id) ON DELETE CASCADE;

CREATE INDEX idx_gavelspaces_tenant ON gavelspaces(tenant_id);
CREATE INDEX idx_projects_tenant ON projects(tenant_id);
CREATE INDEX idx_casefiles_tenant ON casefiles(tenant_id);
CREATE INDEX idx_pleadings_tenant ON pleadings(tenant_id);
