#!/usr/bin/env bash
# Validates docs/ content against README.md design artifacts.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CORE="$ROOT/docs/core-features.md"
INDEX="$ROOT/docs/README.md"
README="$ROOT/README.md"

errors=0
fail() { echo "FAIL: $1" >&2; errors=$((errors + 1)); }
pass() { echo "OK: $1"; }

[[ -f "$CORE" ]] || { fail "missing $CORE"; exit 1; }
[[ -f "$INDEX" ]] || { fail "missing $INDEX"; exit 1; }
[[ -f "$README" ]] || { fail "missing $README"; exit 1; }

# Required sections in core-features.md
sections=(
  "Core concepts"
  "Test Scenario"
  "Agent Set"
  "State Assertion"
  "Architecture"
  "Cluster Provider"
  "Agent Manager"
  "Scenario Engine"
  "Scenario definition format"
  "Failure diagnostics"
  "Fault injection"
  "Scope boundaries"
  "Implementation plan"
  "Technology choices"
)

for section in "${sections[@]}"; do
  if grep -qF "$section" "$CORE"; then
    pass "section: $section"
  else
    fail "missing section heading or title: $section"
  fi
done

# YAML example from README must appear verbatim in core-features.md
yaml_marker='name: scaling-agent-respects-quota-agent'
if grep -qF "$yaml_marker" "$CORE"; then
  pass "YAML example present"
else
  fail "YAML example missing (expected scenario name)"
fi

for yaml_line in \
  'fixtures/namespace-with-quota.yaml' \
  'replicas: 10  # scaling agent wants 10' \
  'value: 5  # quota agent should cap it' \
  'timeout: 120s'; do
  if grep -qF "$yaml_line" "$CORE"; then
    pass "YAML line: $yaml_line"
  else
    fail "YAML example missing line: $yaml_line"
  fi
done

# Fault injection table rows from README
faults=(
  "Kill agent"
  "Network partition"
  "Slow API server"
  "Stale cache"
  "Resource conflict"
)

for fault in "${faults[@]}"; do
  if grep -qF "$fault" "$CORE"; then
    pass "fault row: $fault"
  else
    fail "fault injection table missing: $fault"
  fi
done

# Key README phrases preserved
phrases=(
  "kube-agents-test"
  "eventually consistent"
  "kube-agents-test run scenarios/"
  "Performance/load"
  "No test framework dependency"
)

for phrase in "${phrases[@]}"; do
  if grep -qF "$phrase" "$CORE" || grep -qF "$phrase" "$INDEX"; then
    pass "phrase: $phrase"
  else
    fail "README phrase not found in docs: $phrase"
  fi
done

# docs index references core-features
if grep -qF 'core-features.md' "$INDEX"; then
  pass "docs index links core-features.md"
else
  fail "docs/README.md must link to core-features.md"
fi

if [[ "$errors" -gt 0 ]]; then
  echo "validate-docs.sh: $errors error(s)" >&2
  exit 1
fi

echo "validate-docs.sh: all checks passed"
