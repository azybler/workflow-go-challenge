package workflow

import (
	"context"
	"time"
)

// ExecutionState holds shared state passed between node executors during a workflow run.
type ExecutionState struct {
	FormData  map[string]any
	Condition ConditionInput
	Variables map[string]any // Accumulated outputs (e.g., temperature, conditionResult)
}

// StepResult is the output of executing a single node.
type StepResult struct {
	NodeID   string
	NodeType string
	Label    string
	Status   string         // "completed" or "error"
	Output   map[string]any // Must include "message"; may include type-specific fields
	Duration time.Duration
	Error    string
}

// NodeExecutor defines the interface for executing a single node type.
type NodeExecutor interface {
	Execute(ctx context.Context, node Node, state *ExecutionState) (*StepResult, error)
}

// Registry maps node type strings to their executor implementation.
type Registry map[string]NodeExecutor

// NewRegistry creates a registry populated with all built-in executor types.
func NewRegistry(weatherClient WeatherClient) Registry {
	return Registry{
		"start":       &StartExecutor{},
		"form":        &FormExecutor{},
		"integration": &IntegrationExecutor{client: weatherClient},
		"condition":   &ConditionExecutor{},
		"email":       &EmailExecutor{},
		"end":         &EndExecutor{},
	}
}
