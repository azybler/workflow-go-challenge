# Weather Alert Workflow Backend Implementation

**Date:** 2026-02-16
**Status:** Ready for planning

## What We're Building

Replace the two stubbed Go handlers with a real implementation that:

1. **Persists workflow definitions** (nodes + edges) to PostgreSQL
2. **Executes workflows in-memory** by traversing the node graph, calling the Open-Meteo weather API, evaluating conditions, and returning step-by-step results
3. **Returns results matching the canonical TypeScript `ExecutionResults` type** defined in `web/src/types.ts`

The frontend is complete and working. We are only implementing backend logic.

## Why This Approach

### Database: Single table with JSONB

A single `workflows` table with `id`, `name`, and JSONB columns for `nodes` and `edges`.

**Rationale:**
- Workflows are always loaded and saved as a whole unit (the frontend sends/receives the full graph)
- JSONB preserves the flexible node metadata schema without migrations for each new node type
- New node types (webhooks, transforms, etc.) just add new entries in the JSONB array - no schema changes
- Simple, fast single-row reads - no JOINs needed
- PostgreSQL JSONB supports indexing if we ever need to query into the graph

### Execution Engine: Interface + Registry

```
NodeExecutor interface {
    Execute(ctx, node, state) -> (output, error)
}
```

Concrete executors registered by node type string:
- `StartExecutor` - no-op, marks workflow start
- `FormExecutor` - validates and captures form input
- `IntegrationExecutor` - calls Open-Meteo API with city coordinates
- `ConditionExecutor` - evaluates temperature vs threshold
- `EmailExecutor` - builds mock email payload
- `EndExecutor` - marks workflow completion

**Rationale:**
- Each executor is independently testable with mocked dependencies
- Adding a new node type = implement the interface + register it
- Clean separation of concerns

### Response Format: Canonical TypeScript Types

Return `ExecutionResults` format with `executionId`, `stepNumber`, `nodeType`, `duration`, `output.message`, etc. This matches `web/src/types.ts` exactly.

### Graph Traversal

Walk the graph from the `start` node following edges. For condition nodes, pick the edge whose `sourceHandle` matches the condition result (`"true"` or `"false"`). Stop on error or when reaching an `end` node.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| DB schema | Single `workflows` table + JSONB | Extensible without migrations, matches frontend data shape |
| Execution pattern | `NodeExecutor` interface + type registry | Testable, extensible, clean |
| Response format | Match `ExecutionResults` TypeScript type | Canonical contract, proper typing |
| Weather API | Open-Meteo (free, no API key) | Already specified in node metadata |
| Email | Mock payload in output only | Per requirements - no SMTP |
| Execution persistence | None - in-memory only | Per requirements - config only in DB |
| Testing | Unit tests for executors + integration test for full flow | Good coverage without over-engineering |
| DB migrations | Auto-create table on startup + seed data | Simple for a take-home; mention in NOTES.md |

## Scope

### In scope
- Database schema creation (auto-migrate on startup)
- Seed the sample workflow (ID: `550e8400-e29b-41d4-a716-446655440000`)
- `GET /api/v1/workflows/{id}` - load from DB, return JSON
- `POST /api/v1/workflows/{id}/execute` - parse input, traverse graph, execute nodes, return results
- Open-Meteo API integration with proper error handling and timeouts
- Unit tests for node executors
- Integration tests for workflow execution (mocked HTTP)
- Repository tests
- `NOTES.md` with design rationale and diagrams

### Out of scope (per requirements)
- Real email delivery
- Authentication/authorization
- CRUD endpoints for workflows
- Frontend changes
- CI/CD or deployment

## Open Questions

None - all key decisions resolved.
