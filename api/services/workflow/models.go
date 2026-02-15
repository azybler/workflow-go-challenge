package workflow

import "time"

// Workflow represents a persisted workflow definition with its graph of nodes and edges.
type Workflow struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Nodes     []Node    `json:"nodes"`
	Edges     []Edge    `json:"edges"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Node represents a single step in a workflow graph.
type Node struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Data     NodeData `json:"data"`
}

// Position holds x/y coordinates for rendering the node on the canvas.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NodeData holds the display and configuration data for a node.
type NodeData struct {
	Label       string         `json:"label"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// Edge represents a directed connection between two nodes.
type Edge struct {
	ID           string         `json:"id"`
	Source       string         `json:"source"`
	Target       string         `json:"target"`
	Label        string         `json:"label,omitempty"`
	Type         string         `json:"type,omitempty"`
	SourceHandle string         `json:"sourceHandle,omitempty"`
	TargetHandle string         `json:"targetHandle,omitempty"`
	Animated     bool           `json:"animated,omitempty"`
	Style        map[string]any `json:"style,omitempty"`
	LabelStyle   map[string]any `json:"labelStyle,omitempty"`
}

// ExecuteRequest is the JSON body sent by the frontend to execute a workflow.
type ExecuteRequest struct {
	FormData  map[string]any `json:"formData"`
	Condition ConditionInput `json:"condition"`
}

// ConditionInput holds the operator and threshold for condition evaluation.
type ConditionInput struct {
	Operator  string  `json:"operator"`
	Threshold float64 `json:"threshold"`
}

// ExecutionResults is the top-level response returned after executing a workflow.
type ExecutionResults struct {
	ExecutionID   string         `json:"executionId"`
	Status        string         `json:"status"`
	StartTime     string         `json:"startTime"`
	EndTime       string         `json:"endTime"`
	TotalDuration int64          `json:"totalDuration"`
	Steps         []ExecutionStep `json:"steps"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// ExecutionStep represents the result of executing a single node.
type ExecutionStep struct {
	StepNumber int            `json:"stepNumber"`
	NodeID     string         `json:"nodeId"`
	NodeType   string         `json:"nodeType"`
	Type       string         `json:"type"`
	Label      string         `json:"label"`
	Status     string         `json:"status"`
	Duration   int64          `json:"duration"`
	Output     map[string]any `json:"output"`
	Timestamp  string         `json:"timestamp"`
	Error      string         `json:"error,omitempty"`
}
