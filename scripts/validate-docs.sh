#!/usr/bin/env bash
# QA: ensure docs/core-features.md stays traceable to README.md (design source).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
README="$ROOT/README.md"
CORE="$ROOT/docs/core-features.md"
INDEX="$ROOT/docs/README.md"

fail() { echo "validate-docs: $*" >&2; exit 1; }

[[ -f "$README" ]] || fail "missing README.md"
[[ -f "$CORE" ]] || fail "missing docs/core-features.md"
[[ -f "$INDEX" ]] || fail "missing docs/README.md"

# Required sections (issue #8 checklist)
for heading in \
  "## Core concepts" \
  "## Architecture" \
  "## Scenario definition" \
  "## What this framework tests" \
  "## Failure diagnostics" \
  "## Fault injection" \
  "## Implementation plan" \
  "## Technology choices"
do
  grep -qF "$heading" "$CORE" || fail "core-features.md missing section: $heading"
done

# Index must link to core-features
grep -qF 'core-features.md' "$INDEX" || fail "docs/README.md must link to core-features.md"

# Banned invented API/behavior phrases (not in README)
for phrase in 'dedicated hooks' 'REST API' 'grpc' 'OpenAPI' 'func ' 'interface {'; do
  if grep -qiF "$phrase" "$CORE"; then
    fail "invented or out-of-scope phrase in core-features.md: $phrase"
  fi
done

# Agent Manager wording must match README (kills, not stops)
if grep -qE 'restarts, and stops agents' "$CORE"; then
  fail 'Agent Manager must say "kills" agents per README.md'
fi
grep -qF 'restarts, and kills agents' "$CORE" || fail 'Agent Manager must document kill/restart per README.md'

# Fault injection table rows must match README
while IFS= read -r fault; do
  grep -qF "$fault" "$CORE" || fail "fault injection table missing row: $fault"
done <<'EOF'
Kill agent
Network partition
Slow API server
Stale cache
Resource conflict
EOF

# Tech stack keywords from README
for kw in 'client-go' 'kind' 'standalone binary' 'go test'; do
  grep -qF "$kw" "$CORE" || fail "technology section missing README keyword: $kw"
done

# Example scenario name must match README excerpt
grep -qF 'scaling-agent-respects-quota-agent' "$CORE" "$README" || fail 'scenario example name mismatch'

echo "validate-docs: OK"
