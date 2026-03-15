# ADR-0011: Per-Domain Token Bucket Rate Limiting

## Status
Accepted

## Context
Uncapped request rates burn proxy IPs and trigger platform-side bans.

## Decision
`pkg/platform/ratelimit` implements a per-domain `rate.Limiter` map. Default: 1 req / 30 s per proxy per domain.

## Consequences
Crawl throughput is intentionally limited; configurable via `Config.RateLimitConfig`.
