# ADR-0007: Selector Fallback Chains

## Status
Accepted

## Context
A single dead selector silently returns empty strings; there is no fallback path.

## Decision
Each selector config holds an ordered `[]SelectorRule`. `ExtractField` walks the chain; Prometheus tracks which rule index fires.

## Consequences
Extraction continues on secondary rules; Grafana alert fires when `rule_index > 0` rate spikes.
