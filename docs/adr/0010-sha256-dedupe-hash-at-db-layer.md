# ADR-0010: SHA256 Dedupe Hash at DB Layer

## Status
Accepted

## Context
The same review can be fetched across multiple crawl runs; payload may differ (reply count updates).

## Decision
`dedupe_hash = SHA256(platform + product_id + author_id + reviewed_at_date + review_text[:100])`. `UNIQUE(dedupe_hash)` in both `raw_reviews` and `normalised_reviews`.

## Consequences
Idempotent upserts; deduplication is enforced at DB layer, not application layer.
