# ADR-0012: Session Cookie Cache in Redis

## Status
Accepted

## Context
Each browser session starts cold; platforms detect cold-start fingerprints and serve CAPTCHAs more aggressively.

## Decision
Persist `[]*proto.NetworkCookie` to Redis keyed by `session:{platform}:{proxy_md5}`, TTL 4 h. `BaseScraper.NavigateWithRetry` restores cookies before navigation.

## Consequences
CAPTCHA frequency decreases for returning-session fingerprints; Redis storage is minimal (cookie JSON per slot).
