package workflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testWorkflow() *Workflow {
	return &Workflow{
		ID:   "test-wf",
		Name: "Test Workflow",
		Nodes: []Node{
			{ID: "start", Type: "start", Data: NodeData{Label: "Start"}},
			{ID: "form", Type: "form", Data: NodeData{Label: "User Input"}},
			{
				ID: "weather-api", Type: "integration",
				Data: NodeData{
					Label: "Weather API",
					Metadata: map[string]any{
						"apiEndpoint": "https://api.open-meteo.com/v1/forecast",
						"options": []any{
							map[string]any{"city": "Sydney", "lat": -33.8688, "lon": 151.2093},
						},
					},
				},
			},
			{ID: "condition", Type: "condition", Data: NodeData{Label: "Check Condition"}},
			{
				ID: "email", Type: "email",
				Data: NodeData{
					Label: "Send Alert",
					Metadata: map[string]any{
						"emailTemplate": map[string]any{
							"subject": "Weather Alert",
							"body":    "Alert for {{city}}! Temp: {{temperature}}\u00b0C!",
						},
					},
				},
			},
			{ID: "end", Type: "end", Data: NodeData{Label: "Complete"}},
		},
		Edges: []Edge{
			{ID: "e1", Source: "start", Target: "form"},
			{ID: "e2", Source: "form", Target: "weather-api"},
			{ID: "e3", Source: "weather-api", Target: "condition"},
			{ID: "e4", Source: "condition", Target: "email", SourceHandle: "true"},
			{ID: "e5", Source: "condition", Target: "end", SourceHandle: "false"},
			{ID: "e6", Source: "email", Target: "end"},
		},
	}
}

func TestEngine_HappyPath_ConditionTrue(t *testing.T) {
	client := &mockWeatherClient{temperature: 30.0}
	registry := NewRegistry(client)
	engine := NewEngine(registry)

	state := &ExecutionState{
		FormData:  map[string]any{"name": "Alice", "email": "alice@example.com", "city": "Sydney"},
		Condition: ConditionInput{Operator: "greater_than", Threshold: 25},
		Variables: map[string]any{},
	}

	results, err := engine.Execute(context.Background(), testWorkflow(), state)

	require.NoError(t, err)
	assert.Equal(t, "completed", results.Status)
	assert.Len(t, results.Steps, 6) // start, form, weather, condition, email, end
	assert.NotEmpty(t, results.ExecutionID)
	assert.NotEmpty(t, results.StartTime)
	assert.NotEmpty(t, results.EndTime)

	// Verify step order
	expectedTypes := []string{"start", "form", "integration", "condition", "email", "end"}
	for i, step := range results.Steps {
		assert.Equal(t, expectedTypes[i], step.NodeType, "step %d", i)
		assert.Equal(t, "completed", step.Status)
		assert.Equal(t, i+1, step.StepNumber)
		assert.NotEmpty(t, step.Output["message"])
	}
}

func TestEngine_HappyPath_ConditionFalse(t *testing.T) {
	client := &mockWeatherClient{temperature: 20.0}
	registry := NewRegistry(client)
	engine := NewEngine(registry)

	state := &ExecutionState{
		FormData:  map[string]any{"name": "Bob", "email": "bob@example.com", "city": "Sydney"},
		Condition: ConditionInput{Operator: "greater_than", Threshold: 25},
		Variables: map[string]any{},
	}

	results, err := engine.Execute(context.Background(), testWorkflow(), state)

	require.NoError(t, err)
	assert.Equal(t, "completed", results.Status)
	assert.Len(t, results.Steps, 5) // start, form, weather, condition, end (no email)

	expectedTypes := []string{"start", "form", "integration", "condition", "end"}
	for i, step := range results.Steps {
		assert.Equal(t, expectedTypes[i], step.NodeType, "step %d", i)
		assert.Equal(t, "completed", step.Status)
	}
}

func TestEngine_StopsOnError(t *testing.T) {
	client := &mockWeatherClient{err: fmt.Errorf("API timeout")}
	registry := NewRegistry(client)
	engine := NewEngine(registry)

	state := &ExecutionState{
		FormData:  map[string]any{"name": "Alice", "email": "alice@example.com", "city": "Sydney"},
		Condition: ConditionInput{Operator: "greater_than", Threshold: 25},
		Variables: map[string]any{},
	}

	results, err := engine.Execute(context.Background(), testWorkflow(), state)

	require.NoError(t, err) // Engine returns results, not error
	assert.Equal(t, "failed", results.Status)
	assert.Len(t, results.Steps, 3) // start, form, weather(error)

	assert.Equal(t, "completed", results.Steps[0].Status)
	assert.Equal(t, "completed", results.Steps[1].Status)
	assert.Equal(t, "error", results.Steps[2].Status)
	assert.NotEmpty(t, results.Steps[2].Error)
}

func TestEngine_NoStartNode(t *testing.T) {
	engine := NewEngine(Registry{})

	wf := &Workflow{
		Nodes: []Node{{ID: "end", Type: "end", Data: NodeData{Label: "End"}}},
	}

	_, err := engine.Execute(context.Background(), wf, &ExecutionState{Variables: map[string]any{}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no start node")
}

func TestEngine_CycleProtection(t *testing.T) {
	// Create a simple cycle: a -> b -> a
	registry := Registry{
		"start": &StartExecutor{},
		"end":   &EndExecutor{},
	}
	engine := NewEngine(registry)

	wf := &Workflow{
		Nodes: []Node{
			{ID: "a", Type: "start", Data: NodeData{Label: "A"}},
			{ID: "b", Type: "start", Data: NodeData{Label: "B"}},
		},
		Edges: []Edge{
			{ID: "e1", Source: "a", Target: "b"},
			{ID: "e2", Source: "b", Target: "a"},
		},
	}

	_, err := engine.Execute(context.Background(), wf, &ExecutionState{Variables: map[string]any{}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeded maximum")
}

func TestEngine_UnknownNodeType(t *testing.T) {
	engine := NewEngine(Registry{
		"start": &StartExecutor{},
	})

	wf := &Workflow{
		Nodes: []Node{
			{ID: "start", Type: "start", Data: NodeData{Label: "Start"}},
			{ID: "unknown", Type: "webhook", Data: NodeData{Label: "Webhook"}},
		},
		Edges: []Edge{
			{ID: "e1", Source: "start", Target: "unknown"},
		},
	}

	_, err := engine.Execute(context.Background(), wf, &ExecutionState{Variables: map[string]any{}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no executor registered")
}
