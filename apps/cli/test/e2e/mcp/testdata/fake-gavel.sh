#!/bin/sh
# Fake gavel binary for MCP e2e tests.
# Returns canned responses based on the first argument.

case "$1" in
  judge)
    cat <<'EOF'
{"projects":[{"name":"core","verdict":"pass","findings_count":3,"violations_count":0,"coverage_percent":92.5,"rulings":[{"subtype":"code_quality","passed":true,"detail":"3 findings (0 new)"}]}]}
EOF
    exit 0
    ;;
  validate)
    echo "All checks passed."
    exit 0
    ;;
  config)
    cat <<'EOF'
{"config_path":"/workspace/.gavel/gavel.yaml","gavelspace":"test","projects":[{"name":"core","pattern":"//core/...","languages":["go"],"quality_gate":{"rules":[{"subtype":"code_quality"},{"subtype":"coverage"}]}}]}
EOF
    exit 0
    ;;
  projects)
    cat <<'EOF'
{"projects":[{"name":"core","pattern":"//core/...","languages":["go"],"quality_gate":{"rules":[{"subtype":"code_quality"}]},"baseline":{"findings_count":5,"violations_count":0}}]}
EOF
    exit 0
    ;;
  *)
    echo "unknown command: $1" >&2
    exit 1
    ;;
esac
