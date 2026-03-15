# ADR-0008: Modular Monolith Single Binary

## Status
Accepted

## Context
Microservices add deployment complexity without benefit at current scale.

## Decision
Single Go binary; `cmd/server` for HTTP, `cmd/worker` for Asynq. Packages enforce module boundaries. Split to services later if needed.

## Consequences
Simpler ops; shared memory means no inter-service serialisation overhead for hot paths.
