# ADR-0006: Captcha Service Abstraction

## Status
Accepted

## Context
2Captcha/Anti-Captcha have different APIs and availability windows; switching providers requires code changes.

## Decision
`CaptchaResolver` interface + `Dispatcher` (primary → secondary fallback). Adapters call dispatcher, not providers directly.

## Consequences
Zero adapter-layer changes when swapping providers.
