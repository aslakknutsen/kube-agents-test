#!/usr/bin/env bash
# Adversarial checks: no invented APIs; no cloud providers beyond README.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CORE="$ROOT/docs/core-features.md"
README="$ROOT/README.md"

errors=0
fail() { echo "FAIL: $1" >&2; errors=$((errors + 1)); }
pass() { echo "OK: $1"; }

[[ -f "$CORE" ]] || { fail "missing $CORE"; exit 1; }

# Cloud providers not mentioned in README
for provider in AWS GCP Azure EKS GKE AKS; do
  if grep -qi "$provider" "$CORE"; then
    fail "cloud provider not in README: $provider"
  else
    pass "no cloud provider: $provider"
  fi
done

# Invented Go package paths / HTTP APIs (README has none)
if grep -qE '(package |func |type |interface )[A-Z][a-zA-Z0-9_]*\(' "$CORE"; then
  fail "possible invented Go API signatures in docs"
else
  pass "no Go API signatures"
fi

if grep -qE 'https?://' "$CORE"; then
  fail "URLs in docs (not in README design)"
else
  pass "no HTTP URLs"
fi

# REST paths or gRPC service names not in README
if grep -qE '/api/v[0-9]+/' "$CORE"; then
  fail "invented REST API paths"
else
  pass "no invented REST paths"
fi

# README-allowed cluster backends: kind, k3d, kubeconfig (from diagram/text)
allowed_cluster='kind|k3d|kubeconfig|real cluster'
if grep -qiE 'minikube|microk8s|openshift' "$CORE"; then
  fail "cluster backend not in README"
else
  pass "no extra cluster backends"
fi

# CLI command must match README exactly
if grep -qF 'kube-agents-test run scenarios/' "$CORE"; then
  pass "CLI matches README"
else
  fail "CLI must be exactly: kube-agents-test run scenarios/"
fi

# Framework components from README only
components=(
  "Test Runner"
  "Cluster Provider"
  "Agent Manager"
  "Scenario Engine"
)
for comp in "${components[@]}"; do
  if grep -qF "$comp" "$CORE"; then
    pass "component documented: $comp"
  else
    fail "missing component: $comp"
  fi
done

if [[ "$errors" -gt 0 ]]; then
  echo "validate-docs-adversarial.sh: $errors error(s)" >&2
  exit 1
fi

echo "validate-docs-adversarial.sh: all checks passed"
