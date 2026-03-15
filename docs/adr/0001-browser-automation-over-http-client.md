# ADR-0001: Browser Automation Over HTTP Client

## Status
Accepted

## Context
All three platforms render reviews via JavaScript and fingerprint HTTP clients; token rotation breaks curl-style requests.

## Decision
Use `go-rod` (CDP-based headless Chromium) as the single scraping primitive. No raw HTTP client for review extraction.

## Consequences
Higher resource cost per crawl; mitigated by async job queue and browser pool concurrency limits.
