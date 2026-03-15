# ADR-0005: Two-Stage Pipeline Raw Then Normalise

## Status
Accepted

## Context
Normalisation (language detection, sentiment) is CPU-bound; coupling it to crawl blocks new crawl slots.

## Decision
Scraper writes raw JSON blob → enqueues `normalise:review` task. Normaliser worker runs independently.

## Consequences
Raw data always preserved; re-normalisation after schema change is a backfill, not a re-crawl.
