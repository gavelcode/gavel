-- Seed: default tenant + first admin user.
--
-- Deterministic UUIDs so callers can reference these rows from documentation
-- and CLI scripts. The admin password is the literal "changeme" Argon2id-hashed
-- with a fixed salt; the user is created with must_change_password=true so
-- the first login is forced to set a real password before doing anything else.
--
-- ON CONFLICT DO NOTHING makes this idempotent: Migrate() runs it once on a
-- fresh database, and the testcontainer helper re-applies it after every
-- TRUNCATE to keep test isolation without losing the seed.

INSERT INTO iam_tenants (id, slug, display_name, status, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'default',
    'Default',
    'active',
    '2026-01-01T00:00:00Z'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO iam_users (
    id,
    tenant_id,
    email,
    display_name,
    role,
    password_hash,
    must_change_password,
    is_active,
    created_at
)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'admin@gavel.local',
    'Administrator',
    'admin',
    '$argon2id$v=19$m=65536,t=3,p=4$Z2F2ZWxzZWVkMDAwMDAwMA$ujLGAhG7exjJf+XyAFTNoJjeZYNbk6+aCTdvQ+SZZFk',
    true,
    true,
    '2026-01-01T00:00:00Z'
)
ON CONFLICT (id) DO NOTHING;
