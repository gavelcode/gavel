-- Bootstrap schema for PostgreSQL fresh installs.
-- This file is the single source of truth for the database schema. It is
-- applied once on a brand-new database; subsequent runs are a no-op.

-- All aggregate identifiers are RFC 4122 v4 UUIDs minted by the domain
-- (uuid.New() inside each aggregate factory). Columns that hold them use
-- the native PostgreSQL UUID type — 16 bytes vs. 36 for canonical TEXT,
-- with type-safe binding through pgx/v5. Integer surrogate keys remain
-- TEXT-free identity columns; they are persistence concerns the domain
-- never sees.

-- IAM bounded context (core/domain/iam). These tables back the
-- Vernon-strict User/Session/APIToken/Tenant aggregates that replaced
-- the original platform/identity package.

CREATE TABLE iam_tenants (
    id           UUID PRIMARY KEY,
    slug         TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    status       TEXT NOT NULL CHECK(status IN ('active','suspended')),
    created_at   TIMESTAMPTZ NOT NULL
);

CREATE TABLE iam_users (
    id                   UUID PRIMARY KEY,
    tenant_id            UUID NOT NULL REFERENCES iam_tenants(id) ON DELETE CASCADE,
    email                TEXT NOT NULL,
    display_name         TEXT NOT NULL,
    role                 TEXT NOT NULL CHECK(role IN ('admin','maintainer','viewer')),
    password_hash        TEXT NOT NULL,
    must_change_password BOOLEAN NOT NULL DEFAULT false,
    is_active            BOOLEAN NOT NULL DEFAULT true,
    created_at           TIMESTAMPTZ NOT NULL,
    last_login_at        TIMESTAMPTZ,
    UNIQUE (tenant_id, email)
);

CREATE INDEX idx_iam_users_tenant ON iam_users(tenant_id);

CREATE TABLE iam_sessions (
    id           UUID PRIMARY KEY,
    token_hash   TEXT UNIQUE NOT NULL,
    user_id      UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    user_agent   TEXT NOT NULL DEFAULT '',
    ip_address   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    is_revoked   BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_iam_sessions_user ON iam_sessions(user_id);
CREATE INDEX idx_iam_sessions_expires ON iam_sessions(expires_at);

CREATE TABLE iam_api_tokens (
    id            UUID PRIMARY KEY,
    tenant_id     UUID NOT NULL REFERENCES iam_tenants(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    token_hash    TEXT NOT NULL UNIQUE,
    token_prefix  TEXT NOT NULL,
    scopes        JSONB NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL,
    expires_at    TIMESTAMPTZ,
    last_used_at  TIMESTAMPTZ,
    is_revoked    BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_iam_api_tokens_user ON iam_api_tokens(user_id);
CREATE INDEX idx_iam_api_tokens_tenant ON iam_api_tokens(tenant_id);

CREATE TABLE projects (
    id               UUID PRIMARY KEY,
    key              TEXT NOT NULL UNIQUE,
    name             TEXT NOT NULL,
    target_pattern   TEXT NOT NULL DEFAULT '',
    default_branch   TEXT NOT NULL DEFAULT 'main',
    visibility       TEXT NOT NULL DEFAULT 'private',
    created_by       TEXT,
    created_at       TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')),
    updated_at       TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'))
);

CREATE TABLE project_languages (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    language   TEXT NOT NULL,
    PRIMARY KEY (project_id, language)
);

CREATE TABLE project_quality_gate_rules (
    id              INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    subtype         TEXT NOT NULL,
    strategy_type   TEXT NOT NULL DEFAULT '',
    strategy_params TEXT NOT NULL DEFAULT '',
    sort_order      INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_qg_rules_project ON project_quality_gate_rules(project_id);

CREATE TABLE project_baselines (
    project_id       UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    branch           TEXT NOT NULL,
    fingerprints     JSONB NOT NULL DEFAULT '[]',
    arch_ids         JSONB NOT NULL DEFAULT '[]',
    coverage_percent  DOUBLE PRECISION,
    coverage_by_file  JSONB,
    PRIMARY KEY (project_id, branch)
);

CREATE TABLE casefiles (
    id                   UUID PRIMARY KEY,
    project_id           UUID NOT NULL REFERENCES projects(id),
    commit_sha           TEXT NOT NULL,
    branch               TEXT NOT NULL,
    started_at           TEXT NOT NULL,
    verdict_outcome      TEXT,
    verdict_evaluated_at TEXT,
    total_findings       INTEGER NOT NULL DEFAULT 0,
    new_findings         INTEGER NOT NULL DEFAULT 0,
    existing_findings    INTEGER NOT NULL DEFAULT 0,
    resolved_findings    INTEGER NOT NULL DEFAULT 0,
    is_fresh_evaluation  BOOLEAN NOT NULL DEFAULT false,
    created_at           TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')),
    UNIQUE(project_id, commit_sha)
);

CREATE INDEX idx_casefiles_project ON casefiles(project_id);
CREATE INDEX idx_casefiles_branch ON casefiles(project_id, branch);

CREATE TABLE evidences (
    id           UUID PRIMARY KEY,
    casefile_id  UUID NOT NULL REFERENCES casefiles(id) ON DELETE CASCADE,
    subtype      TEXT NOT NULL,
    source       TEXT NOT NULL,
    collected_at TEXT NOT NULL
);

CREATE INDEX idx_evidences_casefile ON evidences(casefile_id);

CREATE TABLE findings (
    id           INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    evidence_id  UUID NOT NULL REFERENCES evidences(id) ON DELETE CASCADE,
    casefile_id  UUID NOT NULL REFERENCES casefiles(id) ON DELETE CASCADE,
    project_id   UUID NOT NULL REFERENCES projects(id),
    tool         TEXT NOT NULL,
    rule_id      TEXT NOT NULL,
    severity     TEXT NOT NULL,
    file_path    TEXT NOT NULL,
    line         INTEGER NOT NULL DEFAULT 0,
    message      TEXT NOT NULL DEFAULT '',
    fingerprint  TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'new',
    created_at   TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'))
);

CREATE INDEX idx_findings_evidence ON findings(evidence_id);
CREATE INDEX idx_findings_casefile ON findings(casefile_id);
CREATE INDEX idx_findings_project ON findings(project_id);
CREATE INDEX idx_findings_fingerprint ON findings(fingerprint);

CREATE TABLE coverage_data (
    evidence_id   UUID PRIMARY KEY REFERENCES evidences(id) ON DELETE CASCADE,
    total_lines   INTEGER NOT NULL,
    covered_lines INTEGER NOT NULL
);

CREATE TABLE coverage_by_language (
    id            INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    evidence_id   UUID NOT NULL REFERENCES evidences(id) ON DELETE CASCADE,
    language      TEXT NOT NULL,
    total_lines   INTEGER NOT NULL,
    covered_lines INTEGER NOT NULL
);

CREATE TABLE new_code_coverage_data (
    evidence_id     UUID PRIMARY KEY REFERENCES evidences(id) ON DELETE CASCADE,
    covered_lines   INTEGER NOT NULL,
    coverable_lines INTEGER NOT NULL
);

CREATE TABLE architecture_violations (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    evidence_id UUID NOT NULL REFERENCES evidences(id) ON DELETE CASCADE,
    rule        TEXT NOT NULL,
    source_pkg  TEXT NOT NULL,
    target_pkg  TEXT NOT NULL,
    message     TEXT NOT NULL DEFAULT ''
);

CREATE TABLE tool_execution_failures (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    evidence_id UUID NOT NULL REFERENCES evidences(id) ON DELETE CASCADE,
    tool        TEXT NOT NULL,
    reason      TEXT NOT NULL
);

CREATE TABLE rulings (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    casefile_id UUID NOT NULL REFERENCES casefiles(id) ON DELETE CASCADE,
    subtype     TEXT NOT NULL,
    passed      INTEGER NOT NULL,
    detail      TEXT NOT NULL DEFAULT '',
    sort_order  INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_rulings_casefile ON rulings(casefile_id);

CREATE TABLE pleadings (
    id              UUID PRIMARY KEY,
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    number          INTEGER NOT NULL,
    title           TEXT NOT NULL DEFAULT '',
    petitioner      TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'open',
    source_branch   TEXT NOT NULL DEFAULT '',
    target_branch   TEXT NOT NULL DEFAULT '',
    commit_sha      TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')),
    updated_at      TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')),
    UNIQUE(project_id, number)
);

CREATE INDEX idx_pleadings_project ON pleadings(project_id);

CREATE TABLE gavelspaces (
    name       TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')),
    updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'))
);

CREATE TABLE gavelspace_projects (
    gavelspace_name TEXT NOT NULL REFERENCES gavelspaces(name) ON DELETE CASCADE,
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    target_pattern  TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (gavelspace_name, project_id)
);

CREATE TABLE source_blobs (
    project_id    UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    commit_sha    TEXT NOT NULL,
    file_path     TEXT NOT NULL,
    content       BYTEA NOT NULL,
    content_type  TEXT NOT NULL,
    size_bytes    INTEGER NOT NULL,
    created_at    TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')),
    PRIMARY KEY (project_id, commit_sha, file_path)
);

CREATE INDEX idx_source_blobs_commit ON source_blobs(project_id, commit_sha);

CREATE TABLE casefile_file_coverage (
    casefile_id     UUID NOT NULL REFERENCES casefiles(id) ON DELETE CASCADE,
    file_path       TEXT NOT NULL,
    covered_lines   INTEGER[] NOT NULL DEFAULT '{}',
    uncovered_lines INTEGER[] NOT NULL DEFAULT '{}',
    PRIMARY KEY (casefile_id, file_path)
);

CREATE INDEX idx_cfc_casefile ON casefile_file_coverage(casefile_id);
