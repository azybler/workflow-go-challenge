package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles workflow persistence in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{db: pool}
}

// InitSchema creates the workflows table if it does not exist.
func (r *Repository) InitSchema(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS workflows (
			id         UUID PRIMARY KEY,
			name       TEXT NOT NULL DEFAULT '',
			nodes      JSONB NOT NULL DEFAULT '[]',
			edges      JSONB NOT NULL DEFAULT '[]',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

// Seed inserts the sample weather-alert workflow if it does not already exist.
func (r *Repository) Seed(ctx context.Context) error {
	nodesJSON, err := json.Marshal(sampleNodes)
	if err != nil {
		return fmt.Errorf("marshal seed nodes: %w", err)
	}
	edgesJSON, err := json.Marshal(sampleEdges)
	if err != nil {
		return fmt.Errorf("marshal seed edges: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO workflows (id, name, nodes, edges)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, sampleWorkflowID, "Weather Alert Workflow", nodesJSON, edgesJSON)
	if err != nil {
		return fmt.Errorf("seed workflow: %w", err)
	}
	return nil
}

// Get retrieves a workflow by ID. Returns nil, nil if not found.
func (r *Repository) Get(ctx context.Context, id string) (*Workflow, error) {
	var wf Workflow
	var nodesJSON, edgesJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, name, nodes, edges, created_at, updated_at
		FROM workflows WHERE id = $1
	`, id).Scan(&wf.ID, &wf.Name, &nodesJSON, &edgesJSON, &wf.CreatedAt, &wf.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get workflow: %w", err)
	}

	if err := json.Unmarshal(nodesJSON, &wf.Nodes); err != nil {
		return nil, fmt.Errorf("unmarshal nodes: %w", err)
	}
	if err := json.Unmarshal(edgesJSON, &wf.Edges); err != nil {
		return nil, fmt.Errorf("unmarshal edges: %w", err)
	}
	return &wf, nil
}

// InitDB creates the schema and seeds initial data. Called from main on startup.
func InitDB(ctx context.Context, pool *pgxpool.Pool) error {
	repo := NewRepository(pool)
	if err := repo.InitSchema(ctx); err != nil {
		return err
	}
	return repo.Seed(ctx)
}

const sampleWorkflowID = "550e8400-e29b-41d4-a716-446655440000"

var sampleNodes = []Node{
	{
		ID: "start", Type: "start",
		Position: Position{X: -160, Y: 300},
		Data: NodeData{
			Label: "Start", Description: "Begin weather check workflow",
			Metadata: map[string]any{
				"hasHandles": map[string]any{"source": true, "target": false},
			},
		},
	},
	{
		ID: "form", Type: "form",
		Position: Position{X: 152, Y: 304},
		Data: NodeData{
			Label: "User Input", Description: "Process collected data - name, email, location",
			Metadata: map[string]any{
				"hasHandles":      map[string]any{"source": true, "target": true},
				"inputFields":     []string{"name", "email", "city"},
				"outputVariables": []string{"name", "email", "city"},
			},
		},
	},
	{
		ID: "weather-api", Type: "integration",
		Position: Position{X: 460, Y: 304},
		Data: NodeData{
			Label: "Weather API", Description: "Fetch current temperature for {{city}}",
			Metadata: map[string]any{
				"hasHandles":      map[string]any{"source": true, "target": true},
				"inputVariables":  []string{"city"},
				"apiEndpoint":     "https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true",
				"outputVariables": []string{"temperature"},
				"options": []map[string]any{
					{"city": "Sydney", "lat": -33.8688, "lon": 151.2093},
					{"city": "Melbourne", "lat": -37.8136, "lon": 144.9631},
					{"city": "Brisbane", "lat": -27.4698, "lon": 153.0251},
					{"city": "Perth", "lat": -31.9505, "lon": 115.8605},
					{"city": "Adelaide", "lat": -34.9285, "lon": 138.6007},
				},
			},
		},
	},
	{
		ID: "condition", Type: "condition",
		Position: Position{X: 794, Y: 304},
		Data: NodeData{
			Label: "Check Condition", Description: "Evaluate temperature threshold",
			Metadata: map[string]any{
				"hasHandles":          map[string]any{"source": []string{"true", "false"}, "target": true},
				"conditionExpression": "temperature {{operator}} {{threshold}}",
				"outputVariables":     []string{"conditionMet"},
			},
		},
	},
	{
		ID: "email", Type: "email",
		Position: Position{X: 1096, Y: 88},
		Data: NodeData{
			Label: "Send Alert", Description: "Email weather alert notification",
			Metadata: map[string]any{
				"hasHandles":      map[string]any{"source": true, "target": true},
				"inputVariables":  []string{"name", "city", "temperature"},
				"outputVariables": []string{"emailSent"},
				"emailTemplate": map[string]any{
					"subject": "Weather Alert",
					"body":    "Weather alert for {{city}}! Temperature is {{temperature}}\u00b0C!",
				},
			},
		},
	},
	{
		ID: "end", Type: "end",
		Position: Position{X: 1360, Y: 302},
		Data: NodeData{
			Label: "Complete", Description: "Workflow execution finished",
			Metadata: map[string]any{
				"hasHandles": map[string]any{"source": false, "target": true},
			},
		},
	},
}

var sampleEdges = []Edge{
	{ID: "e1", Source: "start", Target: "form", Type: "smoothstep", Animated: true, Style: map[string]any{"stroke": "#10b981", "strokeWidth": 3}, Label: "Initialize"},
	{ID: "e2", Source: "form", Target: "weather-api", Type: "smoothstep", Animated: true, Style: map[string]any{"stroke": "#3b82f6", "strokeWidth": 3}, Label: "Submit Data"},
	{ID: "e3", Source: "weather-api", Target: "condition", Type: "smoothstep", Animated: true, Style: map[string]any{"stroke": "#f97316", "strokeWidth": 3}, Label: "Temperature Data"},
	{ID: "e4", Source: "condition", Target: "email", Type: "smoothstep", SourceHandle: "true", Animated: true, Style: map[string]any{"stroke": "#10b981", "strokeWidth": 3}, Label: "\u2713 Condition Met", LabelStyle: map[string]any{"fill": "#10b981", "fontWeight": "bold"}},
	{ID: "e5", Source: "condition", Target: "end", Type: "smoothstep", SourceHandle: "false", Animated: true, Style: map[string]any{"stroke": "#6b7280", "strokeWidth": 3}, Label: "\u2717 No Alert Needed", LabelStyle: map[string]any{"fill": "#6b7280", "fontWeight": "bold"}},
	{ID: "e6", Source: "email", Target: "end", Type: "smoothstep", Animated: true, Style: map[string]any{"stroke": "#ef4444", "strokeWidth": 2}, Label: "Alert Sent", LabelStyle: map[string]any{"fill": "#ef4444", "fontWeight": "bold"}},
}
