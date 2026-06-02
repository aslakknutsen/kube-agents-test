#!/usr/bin/env bash
# Verify README structure for issue #6 (initiative-driven layout).
set -euo pipefail

README="${1:-README.md}"

if [[ ! -f "$README" ]]; then
  echo "README not found: $README" >&2
  exit 1
fi

fail() {
  echo "verify-readme: $*" >&2
  exit 1
}

grep -q '^## Initiatives' "$README" || fail 'missing ## Initiatives section'

if grep -q '^## Design' "$README"; then
  fail 'top-level ## Design must not exist (content belongs under initiatives)'
fi

grep -q '^## Problem' "$README" || fail 'missing ## Problem section'

initiatives=(
  '### 1. Declarative Scenario Testing'
  '### 2. Cluster & Agent Orchestration'
  '### 3. Multi-Agent Convergence Verification'
  '### 4. Resilience & Fault Injection'
  '### 5. Operability & Delivery'
)

for heading in "${initiatives[@]}"; do
  grep -qF "$heading" "$README" || fail "missing initiative heading: $heading"
done

# Key design tokens preserved from pre-restructure README
tokens=(
  '**Test Scenario**'
  '**Agent Set**'
  '**State Assertion**'
  '**Cluster Provider**'
  '**Agent Manager**'
  '**Scenario Engine**'
  'scaling-agent-respects-quota-agent'
  '| Kill agent |'
  '**Failure diagnostics**'
  '**Implementation plan**'
  '**Tech choices**'
  '| Component | Primary initiative |'
)

for token in "${tokens[@]}"; do
  grep -qF "$token" "$README" || fail "missing design token: $token"
done

echo "verify-readme: OK"
