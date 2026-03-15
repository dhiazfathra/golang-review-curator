# ADR-0002: Selector Config in Database

## Status
Accepted

## Context
UI changes break hardcoded CSS selectors; each change currently requires a code + deploy cycle.

## Decision
Store all selectors in `selector_configs` table with fallback chains. `SelectorStore` hot-reloads every 60 s.

## Consequences
Operator can fix broken selectors with a DB update; no deployment needed for selector changes.
