-- +goose Up

-- Scope name/key uniqueness per tenant. A gavelspace name and a project key are
-- unique WITHIN a tenant, not globally — two tenants may each own a "gavel"
-- gavelspace or a "core" project. Before this, the global gavelspaces PK(name)
-- and projects UNIQUE(key) let one tenant's create silently upsert onto, or be
-- rejected by, another tenant's row. Uniqueness that spans aggregates is a
-- set-based invariant enforced here at the database (Vernon), not in the model.

-- projects: key unique per tenant, not global.
ALTER TABLE projects DROP CONSTRAINT projects_key_key;
ALTER TABLE projects ADD CONSTRAINT projects_tenant_key_key UNIQUE (tenant_id, key);

-- gavelspaces: identity is (tenant_id, name), not name alone. gavelspace_projects
-- references a gavelspace by name, so it gains tenant_id and a composite key.
ALTER TABLE gavelspace_projects ADD COLUMN tenant_id UUID;
UPDATE gavelspace_projects gp
    SET tenant_id = (SELECT g.tenant_id FROM gavelspaces g WHERE g.name = gp.gavelspace_name)
    WHERE tenant_id IS NULL;

ALTER TABLE gavelspace_projects DROP CONSTRAINT gavelspace_projects_gavelspace_name_fkey;
ALTER TABLE gavelspaces DROP CONSTRAINT gavelspaces_pkey;
ALTER TABLE gavelspaces ADD CONSTRAINT gavelspaces_pkey PRIMARY KEY (tenant_id, name);

ALTER TABLE gavelspace_projects ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE gavelspace_projects DROP CONSTRAINT gavelspace_projects_pkey;
ALTER TABLE gavelspace_projects ADD CONSTRAINT gavelspace_projects_pkey
    PRIMARY KEY (tenant_id, gavelspace_name, project_id);
ALTER TABLE gavelspace_projects ADD CONSTRAINT gavelspace_projects_gavelspace_fkey
    FOREIGN KEY (tenant_id, gavelspace_name) REFERENCES gavelspaces(tenant_id, name) ON DELETE CASCADE;
