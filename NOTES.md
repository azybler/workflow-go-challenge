# Weather Alert Workflow Engine -- Design Notes

## Deliverables

- **Workflow persistence:** PostgreSQL-backed storage with auto-schema creation and seed data
- **Workflow execution:** In-memory graph traversal engine with real Open-Meteo weather API integration
- **API endpoints:** `GET /api/v1/workflows/{id}` and `POST /api/v1/workflows/{id}/execute`
- **Tests:** 27 tests covering executors, engine traversal, handlers, and repository
- **Infrastructure fix:** Corrected nginx proxy path stripping bug

## Architecture

```
api/
  main.go                              # Entry point, DB init, HTTP server
  pkg/db/postgres.go                   # PostgreSQL connection pool
  services/workflow/
    models.go                          # Domain types (Workflow, Node, Edge, etc.)
    repository.go                      # DB operations (InitSchema, Seed, Get)
    service.go                         # Service struct, route wiring
    workflow.go                        # HTTP handlers (GET + POST)
    executor.go                        # NodeExecutor interface + Registry
    executors.go                       # 6 concrete executor implementations
    weather.go                         # WeatherClient interface + OpenMeteoClient
    engine.go                          # Graph traversal + execution orchestration
```

### Layering

```
HTTP Handlers  -->  Engine  -->  Executors  -->  External APIs
      |                                              |
      v                                              v
  Repository                                   WeatherClient
      |
      v
  PostgreSQL
```

Handlers parse requests and delegate to the engine. The engine traverses the graph and dispatches each node to its registered executor. Executors are stateless and receive shared state via `ExecutionState`. The weather client is injected as an interface for testability.

## Database Schema (ERD)

```mermaid
erDiagram
    WORKFLOWS {
        uuid id PK "Application-provided UUID"
        text name "Workflow display name"
        jsonb nodes "Array of node objects"
        jsonb edges "Array of edge objects"
        timestamptz created_at "Auto-set on insert"
        timestamptz updated_at "Auto-set on update"
    }
```

Single table with JSONB columns. Schema is created automatically on startup via `CREATE TABLE IF NOT EXISTS`.

## Key Design Decisions

### 1. Single JSONB table for extensibility

**Decision:** Store the entire node/edge graph as JSONB arrays in a single `workflows` table.

**Rationale:**
- Workflows are always loaded and saved as a whole unit -- the frontend sends/receives the full graph
- New node types (webhooks, data transforms, etc.) require zero schema changes -- just add entries to the JSONB array with new `type` values
- Simple, fast single-row reads with no JOINs
- PostgreSQL JSONB supports indexing if we ever need to query into the graph
- Matches the frontend's data model exactly, so serialization is trivial

**Trade-off:** Querying individual nodes across workflows is harder (would need JSONB operators). For this use case, that's acceptable since workflows are always loaded whole.

### 2. NodeExecutor interface + type registry

**Decision:** Define a `NodeExecutor` interface with an `Execute(ctx, node, state)` method. Register concrete executors in a map keyed by node type string.

```go
type NodeExecutor interface {
    Execute(ctx context.Context, node Node, state *ExecutionState) (*StepResult, error)
}
```

**Rationale:**
- **Open/Closed Principle:** Adding a new node type requires only implementing the interface and registering it -- no modification to existing code
- **Testability:** Each executor can be unit tested in isolation with mocked dependencies
- **Clean separation:** Node-specific logic is encapsulated; the engine only knows about the interface

### 3. Graph traversal with edge-based branching

**Decision:** Walk the graph from the start node, following edges. For condition nodes, select the edge whose `sourceHandle` matches the condition result (`"true"` or `"false"`).

**Rationale:**
- Data-driven branching -- the graph structure determines the flow, not hardcoded logic
- Supports arbitrary DAGs, not just linear sequences
- Cycle protection via a 100-step maximum prevents runaway execution

### 4. In-memory execution, no run persistence

**Decision:** Execution runs are ephemeral. Results are computed in-memory and returned to the caller. Only workflow configuration is stored in the database.

**Rationale:** Per requirements. Keeps the system simple and avoids the complexity of execution history storage, cleanup, and querying.

### 5. WeatherClient interface for testability

**Decision:** Extract weather API access behind a `WeatherClient` interface so tests can inject a mock.

**Rationale:** The integration executor calls an external API. Without an interface, tests would need network access, be slow, and be flaky. The mock allows deterministic, fast testing of all execution paths.

### 6. Step status: "completed" vs "success"

**Decision:** Use `"completed"` as the success status for execution steps (not `"success"`).

**Rationale:** The frontend's `ExecutionResults.tsx` component checks `status === 'completed'` to render a green check icon. Using `"success"` (which is what the canonical TypeScript type specifies) would cause the frontend to show a red X for successful steps. Pragmatism over purity.

## Sequence Diagrams

### Happy Path -- Condition Met

```
Client                    Handler              Engine              Executors           Open-Meteo
  |                         |                    |                    |                    |
  |-- POST /execute ------->|                    |                    |                    |
  |   {formData,condition}  |                    |                    |                    |
  |                         |-- Load workflow -->|                    |                    |
  |                         |   from DB (by id)  |                    |                    |
  |                         |                    |                    |                    |
  |                         |-- Execute -------->|                    |                    |
  |                         |                    |-- [start] ------->|                    |
  |                         |                    |<-- completed -----|                    |
  |                         |                    |                    |                    |
  |                         |                    |-- [form] -------->|                    |
  |                         |                    |<-- completed -----|                    |
  |                         |                    |                    |                    |
  |                         |                    |-- [integration] ->|                    |
  |                         |                    |                    |-- GET forecast --->|
  |                         |                    |                    |<-- {temp: 30} -----|
  |                         |                    |<-- completed -----|                    |
  |                         |                    |                    |                    |
  |                         |                    |-- [condition] --->|                    |
  |                         |                    |<-- result=true ---|                    |
  |                         |                    |                    |                    |
  |                         |                    |   Follow edge     |                    |
  |                         |                    |   sourceHandle=   |                    |
  |                         |                    |   "true" -> email |                    |
  |                         |                    |                    |                    |
  |                         |                    |-- [email] ------->|                    |
  |                         |                    |<-- mock draft ----|                    |
  |                         |                    |                    |                    |
  |                         |                    |-- [end] --------->|                    |
  |                         |                    |<-- completed -----|                    |
  |                         |                    |                    |                    |
  |<-- 200 {status:         |<-- results --------|                    |                    |
  |    "completed",         |                    |                    |                    |
  |    steps: [6 steps]}    |                    |                    |                    |
```

### Error Path -- API Failure

```
Client                    Handler              Engine              Executors           Open-Meteo
  |                         |                    |                    |                    |
  |-- POST /execute ------->|                    |                    |                    |
  |                         |-- Execute -------->|                    |                    |
  |                         |                    |-- [start] OK ---->|                    |
  |                         |                    |-- [form] OK ----->|                    |
  |                         |                    |-- [integration] ->|                    |
  |                         |                    |                    |-- GET forecast --->|
  |                         |                    |                    |<-- TIMEOUT --------|
  |                         |                    |<-- error ----------|                    |
  |                         |                    |                    |                    |
  |                         |                    |   STOP execution  |                    |
  |                         |                    |                    |                    |
  |<-- 200 {status:         |<-- partial --------|                    |                    |
  |    "failed",            |    results         |                    |                    |
  |    steps: [3 steps,     |                    |                    |                    |
  |     last has error]}    |                    |                    |                    |
```

Execution errors return HTTP 200 with `status: "failed"` and partial step results. The failed step includes an `error` field. Only request-level errors (bad input, workflow not found) return 4xx/5xx HTTP status codes.

## How to Add a New Node Type

Adding a new node type (e.g., "webhook") requires three steps:

1. **Implement the `NodeExecutor` interface:**

```go
type WebhookExecutor struct {
    client *http.Client
}

func (e *WebhookExecutor) Execute(ctx context.Context, node Node, state *ExecutionState) (*StepResult, error) {
    url, _ := node.Data.Metadata["webhookUrl"].(string)
    // ... call webhook, handle response ...
    return &StepResult{
        NodeID: node.ID, NodeType: node.Type, Label: node.Data.Label,
        Status: "completed",
        Output: map[string]any{"message": "Webhook called successfully"},
    }, nil
}
```

2. **Register it in `NewRegistry()`:**

```go
func NewRegistry(weatherClient WeatherClient) Registry {
    return Registry{
        // ... existing executors ...
        "webhook": &WebhookExecutor{client: &http.Client{Timeout: 10 * time.Second}},
    }
}
```

3. **Add the node to a workflow definition** (in the database or seed data) with `type: "webhook"` and any required metadata.

No schema changes, no migrations, no engine modifications needed.

## Assumptions

- The frontend is the primary consumer; response format prioritizes frontend compatibility
- The Open-Meteo API is free and does not require authentication
- The sample workflow ID `550e8400-e29b-41d4-a716-446655440000` is hardcoded in the frontend
- Workflow definitions are immutable after seeding (no update/delete endpoints)
- Execution runs are ephemeral and not persisted
- City selection is constrained to the 5 cities defined in the integration node's metadata
- Temperature comparisons use 1 decimal place rounding to avoid floating-point issues
- The nginx proxy fix (removing trailing slash from `proxy_pass`) is necessary for production Docker

## Running

```bash
# Start all services
docker-compose up --build

# Run tests (from api/ directory)
cd api && go test ./... -v

# Run with database for repository tests
DATABASE_URL="postgres://workflow:workflow123@localhost:5876/workflow_engine?sslmode=disable" \
  go test ./services/workflow/ -v
```
