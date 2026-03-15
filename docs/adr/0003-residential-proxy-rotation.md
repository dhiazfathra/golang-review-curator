# ADR-0003: Residential Proxy Rotation

## Status
Accepted

## Context
All three platforms IP-ban datacenter ranges; captcha frequency increases without IP diversity.

## Decision
Use BrightData (primary) + Oxylabs (secondary) residential proxy pools with circuit-breaking rotator.

## Consequences
Per-request cost; mitigated by quarantining failing proxies and rate-limiting per domain.
