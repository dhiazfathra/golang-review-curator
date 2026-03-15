# ADR-0009: XHR Interception Primary DOM Fallback

## Status
Accepted

## Context
XHR/GraphQL intercept yields structured JSON; DOM extraction is fragile but always available.

## Decision
XHR intercept is the primary path; DOM extraction via SelectorStore is mandatory fallback. Both paths must be implemented for every adapter.

## Consequences
More code per adapter; rewarded by resilience when XHR path rotates.
