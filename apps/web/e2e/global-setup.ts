import { execFileSync } from "child_process";

const SERVER_PORT = process.env.E2E_SERVER_PORT ?? "3333";
const BASE_URL = `http://localhost:${SERVER_PORT}`;
const PG_CONTAINER = "gavel-e2e-postgres";
const PG_USER = "gavel";
const PG_DB = "gaveltest";

const SEED_EMAIL = "admin@local";
const ADMIN_EMAIL = "admin@local.dev";
const SEED_PASSWORD = "changeme";
const E2E_PASSWORD = "E2eTestPass1!";

const TRUNCATE_DATA_SQL = `
  TRUNCATE gavelspace_projects, gavelspaces,
           pleadings,
           rulings,
           architecture_violations, new_code_coverage_data,
           coverage_by_language, coverage_data,
           findings, evidences, casefiles,
           project_quality_gate_rules, project_languages, projects,
           source_blobs,
           iam_api_tokens, iam_sessions
  CASCADE
`;

async function apiPost(
  path: string,
  body: Record<string, unknown>,
  cookie?: string,
): Promise<{ status: number; body: Record<string, unknown>; cookie?: string }> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (cookie) headers["Cookie"] = cookie;

  const res = await fetch(`${BASE_URL}${path}`, {
    method: "POST",
    headers,
    body: JSON.stringify(body),
  });

  const setCookie = res.headers.get("set-cookie");
  const resBody = res.status === 204 ? {} : await res.json();
  return {
    status: res.status,
    body: resBody as Record<string, unknown>,
    cookie: setCookie?.split(";")[0],
  };
}

function psql(sql: string): void {
  execFileSync(
    "podman",
    ["exec", PG_CONTAINER, "psql", "-U", PG_USER, "-d", PG_DB, "-c", sql],
    { stdio: "pipe" },
  );
}

function resetDatabase(): void {
  psql(TRUNCATE_DATA_SQL.replace(/\n/g, " "));
  psql(`UPDATE iam_users SET email = '${ADMIN_EMAIL}' WHERE email = '${SEED_EMAIL}'`);
}

async function seedTestData(cookie: string): Promise<void> {
  const projectRes = await apiPost(
    "/api/v1/projects",
    {
      key: "core",
      name: "Core Library",
      default_branch: "main",
      target_pattern: "//core/...",
      languages: ["go"],
    },
    cookie,
  );
  if (projectRes.status !== 201) {
    throw new Error(`Failed to create project: ${JSON.stringify(projectRes.body)}`);
  }
  const projectId = projectRes.body.project_id as string;

  const gsRes = await apiPost(
    "/api/v1/gavelspaces",
    { name: "gavel" },
    cookie,
  );
  if (gsRes.status !== 201) {
    throw new Error(`Failed to create gavelspace: ${JSON.stringify(gsRes.body)}`);
  }

  const regRes = await apiPost(
    "/api/v1/gavelspaces/gavel/projects",
    { project_id: projectId, target_pattern: "//core/..." },
    cookie,
  );
  if (regRes.status !== 204) {
    throw new Error(`Failed to register project: ${JSON.stringify(regRes.body)}`);
  }

  const cfRes = await apiPost(
    "/api/v1/casefiles",
    {
      project_id: projectId,
      commit_sha: "e2e1234abc",
      branch: "main",
      commit_author: "e2e-test",
    },
    cookie,
  );
  if (cfRes.status !== 201) {
    throw new Error(`Failed to create casefile: ${JSON.stringify(cfRes.body)}`);
  }
  const casefileId = cfRes.body.case_file_id as string;

  const now = new Date().toISOString();

  const findingsRes = await apiPost(
    `/api/v1/casefiles/${casefileId}/evidence`,
    {
      subtype: "code_quality",
      source: "golangci-lint",
      collected_at: now,
      findings: [
        {
          tool: "golangci-lint",
          rule_id: "errcheck",
          severity: "warning",
          file_path: "core/domain/casefile/model/casefile.go",
          line: 42,
          message: "Error return value not checked",
          fingerprint: "e2e-fp-001",
        },
        {
          tool: "golangci-lint",
          rule_id: "govet",
          severity: "error",
          file_path: "core/domain/project/model/project.go",
          line: 18,
          message: "Possible misuse of unsafe.Pointer",
          fingerprint: "e2e-fp-002",
        },
      ],
    },
    cookie,
  );
  if (findingsRes.status !== 201) {
    throw new Error(`Failed to ingest findings: ${JSON.stringify(findingsRes.body)}`);
  }

  const coverageRes = await apiPost(
    `/api/v1/casefiles/${casefileId}/evidence`,
    {
      subtype: "coverage",
      source: "bazel-coverage",
      collected_at: now,
      coverage: {
        language_stats: [
          { language: "go", lines: 5000, covered_lines: 4250 },
        ],
      },
    },
    cookie,
  );
  if (coverageRes.status !== 201) {
    throw new Error(`Failed to ingest coverage: ${JSON.stringify(coverageRes.body)}`);
  }

  const finalizeRes = await apiPost(
    `/api/v1/casefiles/${casefileId}/finalize`,
    {},
    cookie,
  );
  if (finalizeRes.status !== 200) {
    throw new Error(`Failed to finalize casefile: ${JSON.stringify(finalizeRes.body)}`);
  }

  process.env.E2E_PROJECT_ID = projectId;
  process.env.E2E_CASEFILE_ID = casefileId;
}

async function ensureAdminReady(): Promise<string> {
  const e2eLogin = await apiPost("/api/v1/sessions", {
    email: ADMIN_EMAIL,
    password: E2E_PASSWORD,
  });
  if (e2eLogin.status === 200 && e2eLogin.cookie) {
    console.log("E2E global setup: admin already configured.");
    return e2eLogin.cookie;
  }

  console.log("E2E global setup: first run, logging in with seed password...");
  const seedLogin = await apiPost("/api/v1/sessions", {
    email: ADMIN_EMAIL,
    password: SEED_PASSWORD,
  });
  if (seedLogin.status !== 200 || !seedLogin.cookie) {
    throw new Error(`Login failed with both passwords: ${JSON.stringify(seedLogin.body)}`);
  }

  console.log("E2E global setup: changing admin password...");
  const changePwRes = await apiPost(
    "/api/v1/me/password",
    { current_password: SEED_PASSWORD, new_password: E2E_PASSWORD },
    seedLogin.cookie,
  );
  if (changePwRes.status !== 204) {
    throw new Error(`Password change failed: ${JSON.stringify(changePwRes.body)}`);
  }

  const freshLogin = await apiPost("/api/v1/sessions", {
    email: ADMIN_EMAIL,
    password: E2E_PASSWORD,
  });
  if (freshLogin.status !== 200 || !freshLogin.cookie) {
    throw new Error(`Fresh login failed: ${JSON.stringify(freshLogin.body)}`);
  }
  return freshLogin.cookie;
}

export default async function globalSetup(): Promise<void> {
  console.log("E2E global setup: resetting database...");
  resetDatabase();

  const cookie = await ensureAdminReady();

  console.log("E2E global setup: seeding test data...");
  await seedTestData(cookie);

  console.log("E2E global setup: done.");
}
