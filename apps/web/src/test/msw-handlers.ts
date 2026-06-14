import { http, HttpResponse } from "msw";

export const handlers = [
  http.get("/api/v1/me", () =>
    HttpResponse.json({
      id: 1,
      email: "admin@local",
      display_name: "Admin",
      role: "admin",
      must_change_password: false,
    }),
  ),

  http.post("/api/v1/sessions", async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    if (body.email === "admin@local" && body.password === "admin123!") {
      return HttpResponse.json({
        id: 1,
        email: "admin@local",
        display_name: "Admin",
        role: "admin",
        must_change_password: false,
      });
    }
    return HttpResponse.json({ error: "Invalid credentials" }, { status: 401 });
  }),

  http.delete("/api/v1/sessions/current", () =>
    new HttpResponse(null, { status: 204 }),
  ),

  http.post("/api/v1/me/password", async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    if (body.current_password === "wrong") {
      return HttpResponse.json({ error: "Current password is incorrect" }, { status: 400 });
    }
    return new HttpResponse(null, { status: 204 });
  }),

  http.get("/api/v1/projects", () =>
    HttpResponse.json({
      items: [
        {
          id: "proj-aaa-111", key: "//services/payment/...", name: "Payment",
          default_branch: "main", latest_verdict: "pass", total_findings: 12,
          created_at: "2025-01-01T00:00:00Z",
        },
        {
          id: "proj-bbb-222", key: "//services/auth/...", name: "Auth",
          default_branch: "main", latest_verdict: "", total_findings: 0,
          created_at: "2025-01-02T00:00:00Z",
        },
      ],
      total: 2,
    }),
  ),

  http.get("/api/v1/projects/:key", ({ params }) =>
    HttpResponse.json({
      id: "proj-aaa-111", key: decodeURIComponent(params.key as string), name: "Payment",
      default_branch: "main", latest_verdict: "pass", total_findings: 12,
      created_at: "2025-01-01T00:00:00Z",
      target_pattern: "//services/payment/...",
      languages: ["go", "java"],
      quality_gate_rules: [
        { subtype: "lint_findings", strategy_type: "count_by_severity" },
        { subtype: "coverage", strategy_type: "min_percentage" },
      ],
      severity_counts: { error: 2, warning: 7, note: 3 },
    }),
  ),

  http.post("/api/v1/projects", async () => {
    return HttpResponse.json({ projectId: "proj-ccc-333" }, { status: 201 });
  }),

  http.get("/api/v1/casefiles", ({ request }) => {
    const url = new URL(request.url);
    const projectId = url.searchParams.get("project_id");
    const gavelspace = url.searchParams.get("gavelspace");
    const all = [
      { gavelspace: "alpha", id: "cf-aaa-111", project_id: "proj-aaa-111", commit_sha: "abc1234", branch: "main", started_at: "2025-03-01T09:59:00Z", verdict_outcome: "pass", total_findings: 3, new_findings: 1, existing_findings: 2, resolved_findings: 0, coverage_percent: 82.5, created_at: "2025-03-01T10:00:00Z" },
      { gavelspace: "alpha", id: "cf-bbb-222", project_id: "proj-aaa-111", commit_sha: "def5678", branch: "main", started_at: "2025-02-28T09:59:00Z", verdict_outcome: "fail", total_findings: 5, new_findings: 3, existing_findings: 1, resolved_findings: 1, coverage_percent: 78.3, created_at: "2025-02-28T10:00:00Z" },
      { gavelspace: "beta", id: "cf-ccc-333", project_id: "proj-bbb-222", commit_sha: "beta-sha", branch: "main", started_at: "2025-02-27T09:59:00Z", verdict_outcome: "pass", total_findings: 0, new_findings: 0, existing_findings: 0, resolved_findings: 0, coverage_percent: 90.0, created_at: "2025-02-27T10:00:00Z" },
    ];
    let items = all.map(({ gavelspace: _gs, ...rest }) => rest);
    if (gavelspace) items = all.filter((i) => i.gavelspace === gavelspace).map(({ gavelspace: _gs, ...rest }) => rest);
    if (projectId) items = items.filter((i) => i.project_id === projectId);
    return HttpResponse.json({ items, total: items.length });
  }),

  http.get("/api/v1/casefiles/:id", ({ params }) =>
    HttpResponse.json({
      id: params.id,
      project_id: "proj-aaa-111",
      commit_sha: "abc1234",
      branch: "main",
      started_at: "2025-03-01T09:59:00Z",
      verdict_outcome: "pass",
      total_findings: 3,
      new_findings: 1,
      existing_findings: 2,
      resolved_findings: 0,
      coverage_percent: 82.5,
      created_at: "2025-03-01T10:00:00Z",
    }),
  ),

  http.get("/api/v1/findings", ({ request }) => {
    const url = new URL(request.url);
    const casefileId = url.searchParams.get("casefile_id");
    const severity = url.searchParams.get("severity");
    const tool = url.searchParams.get("tool");
    const filePath = url.searchParams.get("file_path");
    const gavelspace = url.searchParams.get("gavelspace");

    const allItems = [
      { gavelspace: "alpha", tool: "pmd", rule_id: "UnusedVariable", severity: "warning", file_path: "services/payment/handler.go", line: 42, message: "Variable 'x' is never used", fingerprint: "fp1", status: "new", source: "lint", commit_sha: "abc1234", project_key: "alpha-proj" },
      { gavelspace: "alpha", tool: "spotbugs", rule_id: "NP_NULL_ON_SOME_PATH", severity: "error", file_path: "services/auth/login.go", line: 18, message: "Possible null pointer dereference", fingerprint: "fp2", status: "existing", source: "lint", commit_sha: "abc1234", project_key: "alpha-proj" },
      { gavelspace: "beta", tool: "pmd", rule_id: "EmptyCatchBlock", severity: "error", file_path: "services/payment/handler.go", line: 87, message: "Empty catch block", fingerprint: "fp3", status: "new", source: "lint", commit_sha: "def5678", project_key: "beta-proj" },
      { gavelspace: "beta", tool: "errorprone", rule_id: "StringSplitter", severity: "note", file_path: "pkg/utils/parse.go", line: 14, message: "Prefer Splitter to String.split", fingerprint: "fp4", status: "new", source: "lint", commit_sha: "def5678", project_key: "beta-proj" },
    ];

    let items = allItems.map(({ gavelspace: _gs, ...rest }) => rest);
    if (gavelspace) {
      items = allItems.filter((i) => i.gavelspace === gavelspace).map(({ gavelspace: _gs, ...rest }) => rest);
    }
    if (casefileId) items = items.slice(0, 3);
    if (severity) items = items.filter((i) => i.severity === severity);
    if (tool) items = items.filter((i) => i.tool === tool);
    if (filePath) items = items.filter((i) => i.file_path.includes(filePath));

    return HttpResponse.json({ items, total: items.length });
  }),

  http.get("/api/v1/me/tokens", () =>
    HttpResponse.json({
      items: [
        { id: "1", name: "ci-token", prefix: "gav_abc", scopes: ["ingest"], created_at: "2025-01-01T00:00:00Z" },
        { id: "2", name: "read-token", prefix: "gav_def", scopes: ["read"], created_at: "2025-02-01T00:00:00Z", last_used_at: "2025-03-01T00:00:00Z" },
      ],
      next_cursor: null,
    }),
  ),

  http.post("/api/v1/me/tokens", async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      id: "3",
      name: body.name,
      scopes: body.scopes,
      token: "gav_test_full_token_value_here",
      prefix: "gav_tes",
    }, { status: 201 });
  }),

  http.delete("/api/v1/me/tokens/:id", () =>
    new HttpResponse(null, { status: 204 }),
  ),

  http.get("/api/v1/gavelspaces", () =>
    HttpResponse.json({
      items: [
        { name: "acme-corp", project_count: 3, created_at: "2025-01-15T10:00:00Z" },
        { name: "platform", project_count: 1, created_at: "2025-02-01T10:00:00Z" },
      ],
      total: 2,
    }),
  ),

  http.get("/api/v1/gavelspaces/:name", ({ params }) =>
    HttpResponse.json({
      name: params.name,
      projects: [
        { id: "proj-aaa-111", key: "//services/payment/...", name: "Payment", latest_verdict: "pass" },
        { id: "proj-bbb-222", key: "//services/auth/...", name: "Auth", latest_verdict: "fail" },
      ],
      created_at: "2025-01-15T10:00:00Z",
    }),
  ),

  http.post("/api/v1/gavelspaces", async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({ name: body.name }, { status: 201 });
  }),

  http.get("/api/v1/pleadings", ({ request }) => {
    const url = new URL(request.url);
    const projectId = url.searchParams.get("project_id");
    const status = url.searchParams.get("status");
    const gavelspace = url.searchParams.get("gavelspace");
    const all = [
      {
        gavelspace: "alpha",
        id: "pr-aaa-111", project_id: "proj-aaa-111", number: 482,
        title: "Use parameterized queries in posting pipeline",
        petitioner: "Marek Novák", source_branch: "fix/posting-sql", target_branch: "main",
        commit_sha: "e8c1d2f", status: "open", created_at: "2025-03-01T10:00:00Z",
        updated_at: "2025-03-01T10:00:00Z",
        gate_result: {
          passed: true,
          conditions: [
            { label: "No new blocker issues", operator: "<=", value: "0", threshold: "0", passed: true },
            { label: "Coverage on new code", operator: ">=", value: "84.6%", threshold: "80%", passed: true },
          ],
        },
      },
      {
        gavelspace: "beta",
        id: "pr-bbb-222", project_id: "proj-bbb-222", number: 481,
        title: "Add retry logic to payment webhook",
        petitioner: "Ana García", source_branch: "feat/webhook-retry", target_branch: "main",
        commit_sha: "a1b2c3d", status: "open", created_at: "2025-02-28T10:00:00Z",
        updated_at: "2025-02-28T10:00:00Z",
        gate_result: {
          passed: false,
          conditions: [
            { label: "No new blocker issues", operator: "<=", value: "1", threshold: "0", passed: false },
            { label: "Coverage on new code", operator: ">=", value: "72%", threshold: "80%", passed: false },
          ],
        },
      },
    ];
    let items = all.map(({ gavelspace: _gs, ...rest }) => rest);
    if (gavelspace) {
      items = all.filter((i) => i.gavelspace === gavelspace).map(({ gavelspace: _gs, ...rest }) => rest);
    }
    if (projectId) items = items.filter((i) => i.project_id === projectId);
    if (status) items = items.filter((i) => i.status === status);
    return HttpResponse.json({ items, total: items.length });
  }),

  http.get("/api/v1/pleadings/:id", ({ params }) =>
    HttpResponse.json({
      id: params.id, project_id: "proj-aaa-111", number: 482,
      title: "Use parameterized queries in posting pipeline",
      petitioner: "Marek Novák", source_branch: "fix/posting-sql", target_branch: "main",
      commit_sha: "e8c1d2f4a3b", status: "open",
      created_at: "2025-03-01T10:00:00Z", updated_at: "2025-03-01T10:00:00Z",
      gate_result: {
        passed: true,
        conditions: [
          { label: "No new blocker issues", operator: "<=", value: "0", threshold: "0", passed: true },
          { label: "Coverage on new code", operator: ">=", value: "84.6%", threshold: "80%", passed: true },
        ],
      },
    }),
  ),

];
