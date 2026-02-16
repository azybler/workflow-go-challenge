package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const maxSteps = 100

// Engine traverses a workflow graph and executes each node in sequence.
type Engine struct {
	registry Registry
}

// NewEngine creates an Engine with the given executor registry.
func NewEngine(registry Registry) *Engine {
	return &Engine{registry: registry}
}

// Execute traverses the workflow graph starting from the "start" node,
// executing each node via the registry and collecting step results.
// On error, execution stops and partial results are returned with status "failed".
func (e *Engine) Execute(ctx context.Context, wf *Workflow, state *ExecutionState) (*ExecutionResults, error) {
	if state.Variables == nil {
		state.Variables = make(map[string]any)
	}

	startTime := time.Now()

	// Find start node
	current, err := findStartNode(wf.Nodes)
	if err != nil {
		return nil, err
	}

	// Build adjacency: source node ID -> outgoing edges
	edgeMap := buildEdgeMap(wf.Edges)

	// Build node lookup by ID
	nodeMap := make(map[string]*Node, len(wf.Nodes))
	for i := range wf.Nodes {
		nodeMap[wf.Nodes[i].ID] = &wf.Nodes[i]
	}

	var steps []ExecutionStep
	stepNum := 0

	for stepNum < maxSteps {
		executor, ok := e.registry[current.Type]
		if !ok {
			return nil, fmt.Errorf("no executor registered for node type %q", current.Type)
		}

		stepStart := time.Now()
		result, execErr := executor.Execute(ctx, *current, state)
		duration := time.Since(stepStart)

		stepNum++
		step := ExecutionStep{
			StepNumber: stepNum,
			NodeID:     current.ID,
			NodeType:   current.Type,
			Type:       current.Type,
			Label:      current.Data.Label,
			Duration:   duration.Milliseconds(),
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
		}

		if execErr != nil {
			step.Status = "error"
			step.Error = execErr.Error()
			step.Output = map[string]any{"message": fmt.Sprintf("Error: %s", execErr.Error())}
			steps = append(steps, step)

			endTime := time.Now()
			return &ExecutionResults{
				ExecutionID:   uuid.New().String(),
				Status:        "failed",
				StartTime:     startTime.UTC().Format(time.RFC3339),
				EndTime:       endTime.UTC().Format(time.RFC3339),
				TotalDuration: endTime.Sub(startTime).Milliseconds(),
				Steps:         steps,
			}, nil
		}

		step.Status = result.Status
		step.Output = result.Output
		steps = append(steps, step)

		// Find the next node via outgoing edges
		edges := edgeMap[current.ID]
		nextNodeID := ""

		if current.Type == "condition" {
			// For condition nodes, follow the edge whose sourceHandle matches the result
			condResult, _ := state.Variables["conditionResult"].(string)
			for _, edge := range edges {
				if edge.SourceHandle == condResult {
					nextNodeID = edge.Target
					break
				}
			}
		} else if len(edges) > 0 {
			nextNodeID = edges[0].Target
		}

		// No outgoing edge means we've reached a terminal node
		if nextNodeID == "" {
			break
		}

		next, ok := nodeMap[nextNodeID]
		if !ok {
			return nil, fmt.Errorf("edge target node %q not found", nextNodeID)
		}
		current = next
	}

	if stepNum >= maxSteps {
		return nil, fmt.Errorf("execution exceeded maximum of %d steps (possible cycle)", maxSteps)
	}

	endTime := time.Now()
	return &ExecutionResults{
		ExecutionID:   uuid.New().String(),
		Status:        "completed",
		StartTime:     startTime.UTC().Format(time.RFC3339),
		EndTime:       endTime.UTC().Format(time.RFC3339),
		TotalDuration: endTime.Sub(startTime).Milliseconds(),
		Steps:         steps,
	}, nil
}

func findStartNode(nodes []Node) (*Node, error) {
	for i := range nodes {
		if nodes[i].Type == "start" {
			return &nodes[i], nil
		}
	}
	return nil, fmt.Errorf("workflow has no start node")
}

func buildEdgeMap(edges []Edge) map[string][]Edge {
	m := make(map[string][]Edge)
	for _, edge := range edges {
		m[edge.Source] = append(m[edge.Source], edge)
	}
	return m
}
