#!/usr/bin/env bash
#
# devdemo.sh — Self-contained Gavel iterative analysis demo
#
# Demonstrates the full quality gate workflow:
#   baseline → fix arch/security → fix errors → regression blocked → fix dead code → fix errcheck → trends
#
# Requirements: podman, jq, bazel, curl, python3, git
#
set -euo pipefail

# ─── Configuration ───────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
EXAMPLE_DIR="$PROJECT_ROOT/examples/go-repo"

PG_PORT=5432
PG_CONTAINER="gavel-devdemo-postgres"
PG_USER="gavel"
PG_PASS="gavel"
PG_DB="gavel"

SERVER_PORT=8080
ADMIN_EMAIL="admin@gavel.local"
ADMIN_PASS="changeme"
NEW_PASS="DemoP@ss2026"

WORKDIR=""
SERVER_PID=""
COOKIE_JAR=""

# ─── Colors ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

# ─── Helpers ─────────────────────────────────────────────────────────────────
info()   { printf "${BLUE}▸${NC} %s\n" "$*"; }
ok()     { printf "  ${GREEN}✓${NC} %s\n" "$*"; }
warn()   { printf "  ${YELLOW}⚠${NC} %s\n" "$*"; }
fail()   { printf "  ${RED}✗${NC} %s\n" "$*" >&2; exit 1; }
header() { printf "\n${BOLD}━━━ %s ━━━${NC}\n" "$*"; }

run_judge() {
    local stderr_file
    stderr_file=$(mktemp)
    local output
    output=$("$GAVEL_BIN" judge --project api-gateway --json 2>"$stderr_file") || true
    if [ -z "$output" ]; then
        warn "judge produced no JSON output. stderr:"
        cat "$stderr_file" >&2
    fi
    rm -f "$stderr_file"
    echo "$output"
}

jq_field() {
    echo "$1" | jq -r "$2 // empty"
}

assert_eq() {
    local label="$1" expected="$2" actual="$3"
    if [ "$expected" != "$actual" ]; then
        fail "Assertion: $label — expected '$expected', got '$actual'"
    fi
}

assert_ge() {
    local label="$1" min="$2" actual="$3"
    if [ -z "$actual" ]; then
        fail "Assertion: $label — expected >= $min, got empty value"
    fi
    if [ "$actual" -lt "$min" ] 2>/dev/null; then
        fail "Assertion: $label — expected >= $min, got '$actual'"
    fi
}

fix_file() {
    local file="$1"
    shift
    python3 -c "
import sys, pathlib
path = sys.argv[1]
pairs = sys.argv[2:]
p = pathlib.Path(path)
c = p.read_text()
for i in range(0, len(pairs), 2):
    old, new = pairs[i], pairs[i+1]
    if old not in c:
        print(f'ERROR: pattern not found in {path}', file=sys.stderr)
        print(f'  looking for: {repr(old[:80])}...', file=sys.stderr)
        sys.exit(1)
    c = c.replace(old, new, 1)
p.write_text(c)
" "$file" "$@"
}

# ─── Results tracking ────────────────────────────────────────────────────────
SUMMARY_LINES=()

record_cycle() {
    local name="$1" json="$2"
    local findings verdict new_count fixed_count arch cov
    findings=$(jq_field "$json" '.projects[0].findings_count')
    verdict=$(jq_field "$json" '.projects[0].verdict')
    arch=$(jq_field "$json" '.projects[0].violations_count')
    cov=$(jq_field "$json" '.projects[0].coverage_percent' | cut -c1-4)
    new_count=$(jq_field "$json" '.projects[0].delta.new_count')
    fixed_count=$(jq_field "$json" '.projects[0].delta.fixed_count')

    local vcolor="$RED"
    [ "$verdict" = "pass" ] && vcolor="$GREEN"
    printf "  findings=%-4s new=%-3s fixed=%-3s arch=%-2s cov=%s%%  verdict=${vcolor}%s${NC}\n" \
        "${findings:--}" "${new_count:--}" "${fixed_count:--}" "${arch:--}" "${cov:--}" "${verdict:-?}"

    SUMMARY_LINES+=("$(printf "%-28s %8s %5s %5s %5s %7s %7s" \
        "$name" "${findings:--}" "${new_count:--}" "${fixed_count:--}" "${arch:--}" "${cov:--}%" "$verdict")")
}

# ─── Cleanup ─────────────────────────────────────────────────────────────────
cleanup() {
    echo ""
    info "Cleaning up..."
    if [ -n "$SERVER_PID" ]; then
        kill "$SERVER_PID" 2>/dev/null && wait "$SERVER_PID" 2>/dev/null
        ok "Server stopped"
    fi
    if podman rm -f "$PG_CONTAINER" >/dev/null 2>&1; then
        ok "PostgreSQL container removed"
    fi
    if [ -n "$WORKDIR" ] && [ -d "$WORKDIR" ]; then
        chmod -R +w "$WORKDIR" 2>/dev/null || true
        rm -rf "$WORKDIR"
        ok "Workspace removed ($WORKDIR)"
    fi
    [ -n "$COOKIE_JAR" ] && rm -f "$COOKIE_JAR"
}
trap cleanup EXIT

# ═════════════════════════════════════════════════════════════════════════════
printf "${BOLD}"
printf "═══════════════════════════════════════════════════════════════\n"
printf " Gavel — Iterative Analysis Demo (7 cycles)\n"
printf "═══════════════════════════════════════════════════════════════\n"
printf "${NC}"

# ─── Pre-flight checks ──────────────────────────────────────────────────────
header "Pre-flight checks"

for cmd in podman jq bazel curl python3 git rsync; do
    command -v "$cmd" >/dev/null || fail "$cmd not found"
done
ok "Required tools available"

if lsof -i :"$PG_PORT" -sTCP:LISTEN >/dev/null 2>&1; then
    fail "Port $PG_PORT is already in use"
fi
if lsof -i :"$SERVER_PORT" -sTCP:LISTEN >/dev/null 2>&1; then
    fail "Port $SERVER_PORT is already in use"
fi
ok "Ports $PG_PORT and $SERVER_PORT are free"

[ -d "$EXAMPLE_DIR" ] || fail "Example repo not found at $EXAMPLE_DIR"
ok "Example repo found"

# ─── Build binaries ─────────────────────────────────────────────────────────
header "Building binaries"

info "bazel build //apps/cli/cmd/gavel //apps/server/cmd/gavel-server"
cd "$PROJECT_ROOT"
bazel build //apps/cli/cmd/gavel //apps/server/cmd/gavel-server 2>&1 | tail -2

GAVEL_BIN="$PROJECT_ROOT/bazel-bin/apps/cli/cmd/gavel/gavel_/gavel"
SERVER_BIN="$PROJECT_ROOT/bazel-bin/apps/server/cmd/gavel-server/gavel-server_/gavel-server"
[ -x "$GAVEL_BIN" ] || fail "gavel binary not found at $GAVEL_BIN"
[ -x "$SERVER_BIN" ] || fail "gavel-server binary not found at $SERVER_BIN"
ok "Binaries built"

# ─── Start PostgreSQL ────────────────────────────────────────────────────────
header "Starting infrastructure"

podman rm -f "$PG_CONTAINER" >/dev/null 2>&1 || true
podman run -d \
    --name "$PG_CONTAINER" \
    -e POSTGRES_USER="$PG_USER" \
    -e POSTGRES_PASSWORD="$PG_PASS" \
    -e POSTGRES_DB="$PG_DB" \
    -p "$PG_PORT:5432" \
    postgres:16-alpine \
    postgres -c fsync=off -c full_page_writes=off >/dev/null

info "Waiting for PostgreSQL..."
for _ in $(seq 1 30); do
    if podman exec "$PG_CONTAINER" pg_isready -U "$PG_USER" -q 2>/dev/null; then break; fi
    sleep 1
done
podman exec "$PG_CONTAINER" pg_isready -U "$PG_USER" -q 2>/dev/null || fail "PostgreSQL did not start"
ok "PostgreSQL running on port $PG_PORT"

# ─── Start gavel-server ─────────────────────────────────────────────────────
export GAVEL_DATABASE_URL="postgres://$PG_USER:$PG_PASS@localhost:$PG_PORT/$PG_DB?sslmode=disable"
export GAVEL_ADDR=":$SERVER_PORT"

"$SERVER_BIN" serve >/dev/null 2>&1 &
SERVER_PID=$!

info "Waiting for server..."
for _ in $(seq 1 30); do
    if curl -sf "http://localhost:$SERVER_PORT/healthz" >/dev/null 2>&1; then break; fi
    sleep 1
done
curl -sf "http://localhost:$SERVER_PORT/healthz" >/dev/null 2>&1 || fail "Server did not start"
ok "gavel-server running on port $SERVER_PORT (PID $SERVER_PID)"

# ─── Auth bootstrap ──────────────────────────────────────────────────────────
header "Setting up authentication"

COOKIE_JAR=$(mktemp)

curl -sf -X POST "http://localhost:$SERVER_PORT/api/v1/sessions" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}" \
    -c "$COOKIE_JAR" >/dev/null
ok "Logged in as $ADMIN_EMAIL"

curl -sf -X POST "http://localhost:$SERVER_PORT/api/v1/me/password" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR" \
    -d "{\"current_password\":\"$ADMIN_PASS\",\"new_password\":\"$NEW_PASS\"}" >/dev/null
ok "Password changed"

curl -sf -X POST "http://localhost:$SERVER_PORT/api/v1/sessions" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$NEW_PASS\"}" \
    -c "$COOKIE_JAR" >/dev/null
ok "Re-authenticated with new password"

curl -sf -X POST "http://localhost:$SERVER_PORT/api/v1/projects" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR" \
    -d '{"key":"api-gateway","name":"api-gateway","default_branch":"main","target_pattern":"//..."}' >/dev/null
ok "Project api-gateway created"

TOKEN_RESPONSE=$(curl -sf -X POST "http://localhost:$SERVER_PORT/api/v1/me/tokens" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR" \
    -d '{"name":"devdemo","scopes":["ingest"]}')
TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.token')
[ -n "$TOKEN" ] && [ "$TOKEN" != "null" ] || fail "Failed to create API token"
ok "API token: ${TOKEN:0:20}..."

# ─── Prepare workspace ──────────────────────────────────────────────────────
header "Preparing workspace"

WORKDIR=$(mktemp -d)
rsync -a --exclude='bazel-*' --exclude='MODULE.bazel.lock' "$EXAMPLE_DIR/" "$WORKDIR/"
cd "$WORKDIR"

sed -i '' "s|file:///%workspace%/../../hack/registry|file:///$PROJECT_ROOT/hack/registry|" .bazelrc
ok "Registry path fixed → $PROJECT_ROOT/hack/registry"

rm -f .gavel/baseline/api-gateway/{findings,architecture,coverage}
ok "Baselines cleaned"

git init -q
git add -A
git commit -q -m "initial commit"
ok "Git repository initialized at $WORKDIR"

export GAVEL_SERVER_URL="http://localhost:$SERVER_PORT"
export GAVEL_TOKEN="$TOKEN"

# ═════════════════════════════════════════════════════════════════════════════
#  CYCLES
# ═════════════════════════════════════════════════════════════════════════════

# ─── Cycle 1: Baseline ──────────────────────────────────────────────────────
header "Cycle 1: Baseline"

git checkout -q -b demo/cycle-1

info "Running gavel judge (first analysis)..."
RESULT=$(run_judge)
record_cycle "1. Baseline" "$RESULT"

FINDINGS=$(jq_field "$RESULT" '.projects[0].findings_count')
assert_ge "baseline findings" 140 "$FINDINGS"
ok "Baseline established ($FINDINGS findings)"

git add -A
git commit -q -m "chore: establish baseline for api-gateway"
git checkout -q main
git merge -q --no-ff demo/cycle-1 -m "Merge cycle 1: baseline"

# ─── Cycle 2: Architecture + Security ────────────────────────────────────────
header "Cycle 2: Architecture + Security"

git checkout -q -b demo/cycle-2

info "Fixing architecture violations and SQL injection..."

# main.go: remove domain/infrastructure imports
fix_file internal/userinterface/cli/main.go \
    $'\n\t// deliberate archtest: interfaces imports domain directly (bypasses application)\n\t"github.com/example/go-repo/internal/domain/customer"\n\t// deliberate archtest: interfaces imports infrastructure directly\n\t"github.com/example/go-repo/internal/infrastructure/persistence"\n' \
    $'\n'

# main.go: remove direct domain/infrastructure usage
fix_file internal/userinterface/cli/main.go \
    $'\t// deliberate archtest: direct domain usage from interfaces layer\n\tc := customer.Customer{}\n\tfmt.Println("Customer:", c)\n\n\t// deliberate archtest: direct infrastructure usage from interfaces layer\n\trepo := persistence.SQLiteOrderRepo{}\n\tfmt.Println("Repo:", repo)\n\n\tcommand' \
    $'\tcommand'

# sqlite_order_repo.go: remove fmt import + parameterized queries
fix_file internal/infrastructure/persistence/sqlite_order_repo.go \
    $'\t"database/sql"\n\t"fmt"' \
    $'\t"database/sql"'

fix_file internal/infrastructure/persistence/sqlite_order_repo.go \
    $'\t// deliberate gosec G201: SQL injection via fmt.Sprintf\n\tquery := fmt.Sprintf("INSERT INTO orders (id, customer_id, status) VALUES (%d, %d, \'%s\')", o.ID(), o.CustomerID(), o.Status())\n\t_, err := r.db.Exec(query)' \
    $'\t_, err := r.db.Exec("INSERT INTO orders (id, customer_id, status) VALUES (?, ?, ?)", o.ID(), o.CustomerID(), o.Status())'

fix_file internal/infrastructure/persistence/sqlite_order_repo.go \
    $'\tquery := fmt.Sprintf("SELECT id, customer_id FROM orders WHERE id = %d", id)\n\trow := r.db.QueryRow(query)' \
    $'\trow := r.db.QueryRow("SELECT id, customer_id FROM orders WHERE id = ?", id)'

# sqlite_customer_repo.go: remove fmt import + parameterized query
fix_file internal/infrastructure/persistence/sqlite_customer_repo.go \
    $'\t"database/sql"\n\t"fmt"' \
    $'\t"database/sql"'

fix_file internal/infrastructure/persistence/sqlite_customer_repo.go \
    $'\tquery := fmt.Sprintf("SELECT id, name, email FROM customers WHERE id = %d", id)\n\trow := r.db.QueryRow(query)' \
    $'\trow := r.db.QueryRow("SELECT id, name, email FROM customers WHERE id = ?", id)'

git add -A
git commit -q -m "fix: resolve architecture violations and SQL injection"

info "Running gavel judge..."
RESULT=$(run_judge)
record_cycle "2. Arch + Security" "$RESULT"

assert_eq "no new findings" "0" "$(jq_field "$RESULT" '.projects[0].delta.new_count')"
ok "Architecture and security fixed"

git checkout -q main
git merge -q --no-ff demo/cycle-2 -m "Merge cycle 2: architecture and security fixes"

# ─── Cycle 3: Error handling ─────────────────────────────────────────────────
header "Cycle 3: Error Handling"

git checkout -q -b demo/cycle-3

info "Fixing defer-in-loop, errcheck on rows.Close, errcheck in order tests..."

# sqlite_customer_repo.go: move defer out of loop + handle rows.Close
fix_file internal/infrastructure/persistence/sqlite_customer_repo.go \
    $'\tvar customers []customer.Customer\n\t// deliberate gocritic: defer in loop\n\tfor rows.Next() {\n\t\tdefer rows.Close()' \
    $'\tdefer func() { _ = rows.Close() }()\n\tvar customers []customer.Customer\n\tfor rows.Next() {'

# order_test.go: check error on Confirm (line 70)
fix_file internal/domain/order/order_test.go \
    $'\to.Confirm()\n\n\tif err := o.Confirm(); err == nil {\n\t\tt.Fatal("expected error confirming already confirmed order")' \
    $'\tif err := o.Confirm(); err != nil {\n\t\tt.Fatalf("unexpected error on first confirm: %v", err)\n\t}\n\n\tif err := o.Confirm(); err == nil {\n\t\tt.Fatal("expected error confirming already confirmed order")'

# order_test.go: check error on Confirm + MarkPaid (lines 82-83)
fix_file internal/domain/order/order_test.go \
    $'\to.Confirm()\n\to.MarkPaid()\n\n\tif o.Status() != order.StatusPaid {' \
    $'\tif err := o.Confirm(); err != nil {\n\t\tt.Fatalf("unexpected error confirming: %v", err)\n\t}\n\tif err := o.MarkPaid(); err != nil {\n\t\tt.Fatalf("unexpected error marking paid: %v", err)\n\t}\n\n\tif o.Status() != order.StatusPaid {'

# order_test.go: check error on Cancel (line 92)
fix_file internal/domain/order/order_test.go \
    $'\to.Cancel()\n\tif o.Status() != order.StatusCancelled {' \
    $'\tif err := o.Cancel(); err != nil {\n\t\tt.Fatalf("unexpected error cancelling: %v", err)\n\t}\n\tif o.Status() != order.StatusCancelled {'

git add -A
git commit -q -m "fix: resolve error handling findings (defer-in-loop, errcheck)"

info "Running gavel judge..."
RESULT=$(run_judge)
record_cycle "3. Error Handling" "$RESULT"

assert_eq "no new findings" "0" "$(jq_field "$RESULT" '.projects[0].delta.new_count')"
ok "Error handling fixed"

git checkout -q main
git merge -q --no-ff demo/cycle-3 -m "Merge cycle 3: error handling fixes"

# ─── Cycle 4: Regression (blocked by gate) ───────────────────────────────────
header "Cycle 4: Regression (blocked by gate)"

git checkout -q -b demo/cycle-4

info "Introducing 2 new findings (forbidigo + errcheck)..."

# Append a function with deliberate violations to main.go
python3 -c "
import pathlib
p = pathlib.Path('internal/userinterface/cli/main.go')
c = p.read_text()
c += '''
func debugCheck() {
\tfmt.Println(\"DEBUG: system check\")
\tos.Getwd()
}
'''
p.write_text(c)
"

git add -A
git commit -q -m "feat: add debug check (introduces new findings)"

info "Running gavel judge (should be BLOCKED)..."
RESULT=$(run_judge)
NEW_COUNT=$(jq_field "$RESULT" '.projects[0].delta.new_count')
printf "  ${RED}BLOCKED${NC}: new_count=%s (gate rejects new findings)\n" "$NEW_COUNT"
assert_ge "regression detected" 1 "$NEW_COUNT"
record_cycle "4a. Regression (BLOCKED)" "$RESULT"
ok "Gate correctly blocked the regression"

info "Reverting bad code..."
git checkout -q HEAD~1 -- internal/userinterface/cli/main.go
git add -A
git commit -q -m "fix: remove debug check, resolve regression"

info "Running gavel judge (should pass code_quality)..."
RESULT=$(run_judge)
record_cycle "4. Regression fixed" "$RESULT"

assert_eq "no new findings after fix" "0" "$(jq_field "$RESULT" '.projects[0].delta.new_count')"
ok "Regression resolved — gate passes"

git checkout -q main
git merge -q --no-ff demo/cycle-4 -m "Merge cycle 4: regression introduced and fixed"

# ─── Cycle 5: Dead code + useless assignments ────────────────────────────────
header "Cycle 5: Dead Code + Useless Assignments"

git checkout -q -b demo/cycle-5

info "Fixing unreachable code, unused type, ineffective assignment..."

# address.go: remove unreachable return
fix_file internal/domain/customer/address.go \
    $'\treturn a.street + ", " + a.city\n\treturn a.street + ", " + a.city + " " + a.zipCode // deliberate unreachable code' \
    $'\treturn a.street + ", " + a.city + " " + a.zipCode'

# product.go: remove unused lowercase product struct
fix_file internal/domain/product/product.go \
    $'// deliberate revive: exported field in unexported struct\ntype product struct {\n\tID    int\n\tname  string\n\tdesc  string\n\tprice order.Money\n}\n\ntype Product' \
    $'type Product'

# database_config.go: remove dead maxConns block
fix_file internal/platform/config/database_config.go \
    $'\n\tmaxConns := 10\n\tif maxConns > 25 {\n\t\tmaxConns = 25\n\t}\n\n\treturn' \
    $'\n\treturn'

git add -A
git commit -q -m "fix: remove dead code and useless assignments"

info "Running gavel judge..."
RESULT=$(run_judge)
record_cycle "5. Dead Code" "$RESULT"

assert_eq "no new findings" "0" "$(jq_field "$RESULT" '.projects[0].delta.new_count')"
ok "Dead code removed"

git checkout -q main
git merge -q --no-ff demo/cycle-5 -m "Merge cycle 5: dead code and useless assignments"

# ─── Cycle 6: Errcheck in payment tests ──────────────────────────────────────
header "Cycle 6: Errcheck in Payment Tests"

git checkout -q -b demo/cycle-6

info "Fixing unchecked errors in payment_test.go..."

# TestProcessAndFail: check Process + Fail
fix_file internal/domain/payment/payment_test.go \
    $'\tp.Process()\n\tp.Fail()' \
    $'\tif err := p.Process(); err != nil {\n\t\tt.Fatalf("process: %v", err)\n\t}\n\tif err := p.Fail(); err != nil {\n\t\tt.Fatalf("fail: %v", err)\n\t}'

# TestRefundCompleted: check Process + Complete
fix_file internal/domain/payment/payment_test.go \
    $'\tp.Process()\n\tp.Complete()\n\tif err := p.Refund()' \
    $'\tif err := p.Process(); err != nil {\n\t\tt.Fatalf("process: %v", err)\n\t}\n\tif err := p.Complete(); err != nil {\n\t\tt.Fatalf("complete: %v", err)\n\t}\n\tif err := p.Refund()'

# TestCannotProcessNonPending: check first Process
fix_file internal/domain/payment/payment_test.go \
    $'\tp.Process()\n\tif err := p.Process(); err == nil {' \
    $'\tif err := p.Process(); err != nil {\n\t\tt.Fatalf("first process: %v", err)\n\t}\n\tif err := p.Process(); err == nil {'

git add -A
git commit -q -m "fix: resolve errcheck findings in payment tests"

info "Running gavel judge..."
RESULT=$(run_judge)
record_cycle "6. Payment Errcheck" "$RESULT"

assert_eq "no new findings" "0" "$(jq_field "$RESULT" '.projects[0].delta.new_count')"
ok "Payment errcheck resolved"

git checkout -q main
git merge -q --no-ff demo/cycle-6 -m "Merge cycle 6: errcheck in payment tests"

# ─── Cycle 7: Trends ─────────────────────────────────────────────────────────
header "Cycle 7: Trends"

info "Fetching trends from server..."
echo ""
"$GAVEL_BIN" trends --project api-gateway 2>/dev/null || true
echo ""

# ═════════════════════════════════════════════════════════════════════════════
#  SUMMARY
# ═════════════════════════════════════════════════════════════════════════════
header "Summary"

printf "${BOLD}%-28s %8s %5s %5s %5s %7s %7s${NC}\n" \
    "Cycle" "Findngs" "New" "Fixed" "Arch" "Cov" "Verdict"
printf "%-28s %8s %5s %5s %5s %7s %7s\n" \
    "────────────────────────────" "────────" "─────" "─────" "─────" "───────" "───────"
for line in "${SUMMARY_LINES[@]}"; do
    echo "$line"
done

echo ""
ok "Demo complete! All cycles executed successfully."
