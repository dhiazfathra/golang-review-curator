#!/usr/bin/env bash
# =============================================================================
# scripts/coverage.sh — CI coverage script aligned with codecov.yml
#
# Produces a single coverage.out that Codecov's uploader consumes.
# Exclusions here mirror the `ignore:` block in codecov.yml exactly.
#
# Usage (locally):    bash scripts/coverage.sh
# Usage (CI):         bash scripts/coverage.sh --upload
#
# Flags:
#   --upload          Upload to Codecov after generating the report
#   --check           Exit non-zero if total < COVERAGE_THRESHOLD
#   --html            Also generate an HTML report
#   --race            Enable -race detector (default: on)
#   --no-race         Disable -race detector
#   --timeout <d>     Override test timeout (default: 5m)
#   --threshold <n>   Override minimum coverage % (default: 80)
# =============================================================================
set -euo pipefail

# ---------------------------------------------------------------------------
# Defaults (override via flags or environment variables)
# ---------------------------------------------------------------------------
COVERAGE_OUT="${COVERAGE_OUT:-coverage.out}"
COVERAGE_HTML="${COVERAGE_HTML:-coverage.html}"
COVERAGE_THRESHOLD="${COVERAGE_THRESHOLD:-80}"   # Must match codecov.yml target
GOTEST_TIMEOUT="${GOTEST_TIMEOUT:-5m}"
RACE="-race"
DO_UPLOAD=false
DO_CHECK=false
DO_HTML=false

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    --upload)    DO_UPLOAD=true  ;;
    --check)     DO_CHECK=true   ;;
    --html)      DO_HTML=true    ;;
    --race)      RACE="-race"    ;;
    --no-race)   RACE=""         ;;
    --timeout)   GOTEST_TIMEOUT="$2"; shift ;;
    --threshold) COVERAGE_THRESHOLD="$2"; shift ;;
    *) echo "Unknown flag: $1" >&2; exit 1 ;;
  esac
  shift
done

# ---------------------------------------------------------------------------
# Resolve module path
# ---------------------------------------------------------------------------
MODULE=$(go list -m)
echo "==> Module: ${MODULE}"

# ---------------------------------------------------------------------------
# EXCLUDE_PATTERN — mirrors codecov.yml `ignore:` list
#
# codecov.yml ignore entry          grep -Ev pattern
# ─────────────────────────────────────────────────────
# **/*_test.go                      (Go never instruments; skip)
# **/mock_*.go                      /mock[^/]*\.go   (filename match)
# **/mocks/**                       /mocks/          (directory match)
# **/testdata/**                    (Go skips by convention; skip)
# **/*.pb.go                        \.pb\.go$
# **/*.gen.go                       \.gen\.go$
# **/vendor/**                      (go list never returns vendor; skip)
# cmd/*/main.go                     ${MODULE}/cmd/   (import-path match)
# ---------------------------------------------------------------------------
EXCLUDE_PATTERN="/mock[^/]*\.go|/mocks/|\.pb\.go$|\.gen\.go$|${MODULE}/cmd/"

# ---------------------------------------------------------------------------
# Build package list (import paths, not file globs)
# ---------------------------------------------------------------------------
echo "==> Resolving packages (excluding: ${EXCLUDE_PATTERN})"
ALL_PKGS=$(go list ./... | grep -Ev "${EXCLUDE_PATTERN}" | tr '\n' ' ')

if [[ -z "${ALL_PKGS}" ]]; then
  echo "ERROR: No packages found after exclusions." >&2
  exit 1
fi

COVERPKG=$(echo "${ALL_PKGS}" | tr ' ' ',')
echo "==> Packages to test & instrument:"
echo "${ALL_PKGS}" | tr ' ' '\n' | sed 's/^/    /'

# ---------------------------------------------------------------------------
# Run tests
# ---------------------------------------------------------------------------
echo ""
echo "==> Running go test (timeout: ${GOTEST_TIMEOUT}, race: ${RACE:-off})"
# shellcheck disable=SC2086  # intentional word splitting for $RACE and $ALL_PKGS
go test \
  ${RACE} \
  -timeout "${GOTEST_TIMEOUT}" \
  -covermode=atomic \
  -coverprofile="${COVERAGE_OUT}" \
  -coverpkg="${COVERPKG}" \
  ${ALL_PKGS}

# ---------------------------------------------------------------------------
# Post-process: strip generated/excluded files from coverage.out
# This makes `go tool cover -func` totals match what Codecov reports.
# Mirrors the same patterns as the grep above but at the file level.
# ---------------------------------------------------------------------------
echo ""
echo "==> Stripping excluded files from ${COVERAGE_OUT}"
# Use scripts/strip_coverage.awk to avoid inline character class issues
# (e.g. [^/] being misinterpreted when awk is invoked via shell heredoc or
# backslash-newline continuation).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
awk -f "${SCRIPT_DIR}/strip_coverage.awk" "${COVERAGE_OUT}" > "${COVERAGE_OUT}.tmp"
mv "${COVERAGE_OUT}.tmp" "${COVERAGE_OUT}"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "==> Coverage summary"
go tool cover -func="${COVERAGE_OUT}" | grep -E "^total:|100\.0%"

TOTAL=$(go tool cover -func="${COVERAGE_OUT}" \
  | awk '/^total:/ { gsub(/%/, "", $NF); printf "%d", $NF }')
echo ""
echo "    Total: ${TOTAL}%  (threshold: ${COVERAGE_THRESHOLD}%)"

# ---------------------------------------------------------------------------
# Optional: HTML report
# ---------------------------------------------------------------------------
if [[ "${DO_HTML}" == "true" ]]; then
  echo ""
  echo "==> Generating HTML report → ${COVERAGE_HTML}"
  go tool cover -html="${COVERAGE_OUT}" -o "${COVERAGE_HTML}"
fi

# ---------------------------------------------------------------------------
# Optional: threshold check
# ---------------------------------------------------------------------------
if [[ "${DO_CHECK}" == "true" ]]; then
  echo ""
  if [[ "${TOTAL}" -lt "${COVERAGE_THRESHOLD}" ]]; then
    echo "FAIL: ${TOTAL}% is below the required threshold of ${COVERAGE_THRESHOLD}%"
    exit 1
  else
    echo "PASS: ${TOTAL}% meets the required threshold of ${COVERAGE_THRESHOLD}%"
  fi
fi

# ---------------------------------------------------------------------------
# Optional: upload to Codecov
# ---------------------------------------------------------------------------
if [[ "${DO_UPLOAD}" == "true" ]]; then
  echo ""
  echo "==> Uploading to Codecov"
  if command -v codecov &>/dev/null; then
    # Codecov CLI (v0.6+)
    codecov upload-process \
      --file "${COVERAGE_OUT}" \
      --plugin noop           # noop: coverage.out already captured above
      # --flag unit           # Uncomment to set a flag (mirrors codecov.yml flags)
      # --slug owner/repo     # Required outside GitHub Actions
  elif command -v curl &>/dev/null; then
    # Fallback: bash uploader (legacy, not recommended for production)
    bash <(curl -s https://codecov.io/bash) -f "${COVERAGE_OUT}"
  else
    echo "WARNING: Neither 'codecov' CLI nor 'curl' found. Skipping upload." >&2
  fi
fi

echo ""
echo "==> Done. Report: ${COVERAGE_OUT}"
