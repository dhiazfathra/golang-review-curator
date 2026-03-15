# ADR-0004: Asynq for Crawl Queue

## Status
Accepted

## Context
Crawl jobs are long-running, retryable, and need priority lanes (crawl vs normalise).

## Decision
Use `asynq` (Redis-backed) for job queue. Queues: `crawl` (priority 2) and `normalise` (priority 1).

## Consequences
Redis becomes a required infrastructure dependency; Asynq dashboard available for ops visibility.
