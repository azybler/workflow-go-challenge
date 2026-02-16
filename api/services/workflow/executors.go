package workflow

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// StartExecutor handles the "start" node type. It is a no-op that marks the workflow beginning.
type StartExecutor struct{}

func (e *StartExecutor) Execute(_ context.Context, node Node, _ *ExecutionState) (*StepResult, error) {
	return &StepResult{
		NodeID: node.ID, NodeType: node.Type, Label: node.Data.Label,
		Status: "completed",
		Output: map[string]any{"message": "Workflow execution started"},
	}, nil
}

// FormExecutor handles the "form" node type. It captures and validates user input.
type FormExecutor struct{}

func (e *FormExecutor) Execute(_ context.Context, node Node, state *ExecutionState) (*StepResult, error) {
	for _, field := range []string{"name", "email", "city"} {
		val, ok := state.FormData[field]
		if !ok {
			return nil, fmt.Errorf("missing required field: %s", field)
		}
		if s, ok := val.(string); !ok || strings.TrimSpace(s) == "" {
			return nil, fmt.Errorf("field %s must be a non-empty string", field)
		}
	}

	return &StepResult{
		NodeID: node.ID, NodeType: node.Type, Label: node.Data.Label,
		Status: "completed",
		Output: map[string]any{
			"message":  fmt.Sprintf("Collected user input for %s", state.FormData["name"]),
			"formData": state.FormData,
		},
	}, nil
}

// IntegrationExecutor handles the "integration" node type. It calls an external weather API.
type IntegrationExecutor struct {
	client WeatherClient
}

func (e *IntegrationExecutor) Execute(ctx context.Context, node Node, state *ExecutionState) (*StepResult, error) {
	city, _ := state.FormData["city"].(string)

	// Look up coordinates from node metadata options
	options, _ := node.Data.Metadata["options"].([]any)
	var lat, lon float64
	var found bool
	for _, opt := range options {
		m, ok := opt.(map[string]any)
		if !ok {
			continue
		}
		if cityName, _ := m["city"].(string); strings.EqualFold(cityName, city) {
			var okLat, okLon bool
			lat, okLat = toFloat64(m["lat"])
			lon, okLon = toFloat64(m["lon"])
			if !okLat || !okLon {
				return nil, fmt.Errorf("invalid coordinates for city %q", city)
			}
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("city %q not found in available options", city)
	}

	endpoint, _ := node.Data.Metadata["apiEndpoint"].(string)

	temperature, err := e.client.GetTemperature(ctx, lat, lon)
	if err != nil {
		return nil, fmt.Errorf("weather API error: %w", err)
	}

	state.Variables["temperature"] = temperature

	return &StepResult{
		NodeID: node.ID, NodeType: node.Type, Label: node.Data.Label,
		Status: "completed",
		Output: map[string]any{
			"message":     fmt.Sprintf("Current temperature in %s: %.1f\u00b0C", city, temperature),
			"temperature": temperature,
			"location":    city,
			"apiResponse": map[string]any{
				"endpoint":   endpoint,
				"method":     "GET",
				"statusCode": 200,
				"data":       map[string]any{"temperature": temperature},
			},
		},
	}, nil
}

// ConditionExecutor handles the "condition" node type. It evaluates a temperature comparison.
type ConditionExecutor struct{}

func (e *ConditionExecutor) Execute(_ context.Context, node Node, state *ExecutionState) (*StepResult, error) {
	tempRaw, ok := state.Variables["temperature"]
	if !ok {
		return nil, fmt.Errorf("temperature variable not set")
	}
	temperature, ok := toFloat64(tempRaw)
	if !ok {
		return nil, fmt.Errorf("temperature is not a number")
	}

	operator := state.Condition.Operator
	threshold := state.Condition.Threshold
	result := evaluateCondition(temperature, operator, threshold)

	if result {
		state.Variables["conditionResult"] = "true"
	} else {
		state.Variables["conditionResult"] = "false"
	}

	symbol := operatorSymbol(operator)
	expression := fmt.Sprintf("%.1f %s %.1f", temperature, symbol, threshold)

	var message string
	if result {
		message = fmt.Sprintf("Temperature %.1f\u00b0C is %s %.1f\u00b0C - condition met", temperature, operatorLabel(operator), threshold)
	} else {
		message = fmt.Sprintf("Temperature %.1f\u00b0C is not %s %.1f\u00b0C - condition not met", temperature, operatorLabel(operator), threshold)
	}

	return &StepResult{
		NodeID: node.ID, NodeType: node.Type, Label: node.Data.Label,
		Status: "completed",
		Output: map[string]any{
			"message":      message,
			"conditionMet": result,
			"conditionResult": map[string]any{
				"expression":  expression,
				"result":      result,
				"temperature": temperature,
				"operator":    operator,
				"threshold":   threshold,
			},
		},
	}, nil
}

// EmailExecutor handles the "email" node type. It produces a mock email payload.
type EmailExecutor struct{}

func (e *EmailExecutor) Execute(_ context.Context, node Node, state *ExecutionState) (*StepResult, error) {
	name, _ := state.FormData["name"].(string)
	email, _ := state.FormData["email"].(string)
	city, _ := state.FormData["city"].(string)
	temperature, _ := state.Variables["temperature"].(float64)

	tmpl, _ := node.Data.Metadata["emailTemplate"].(map[string]any)
	subject, _ := tmpl["subject"].(string)
	body, _ := tmpl["body"].(string)

	// Template substitution
	replacer := strings.NewReplacer(
		"{{name}}", name,
		"{{city}}", city,
		"{{temperature}}", fmt.Sprintf("%.1f", temperature),
	)
	body = replacer.Replace(body)
	subject = replacer.Replace(subject)

	emailDraft := map[string]any{
		"to":        email,
		"from":      "weather-alerts@example.com",
		"subject":   subject,
		"body":      body,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	return &StepResult{
		NodeID: node.ID, NodeType: node.Type, Label: node.Data.Label,
		Status: "completed",
		Output: map[string]any{
			"message":    fmt.Sprintf("Weather alert email drafted for %s", email),
			"emailDraft": emailDraft,
			"emailContent": map[string]any{
				"to":      email,
				"subject": subject,
				"body":    body,
			},
			"emailSent": true,
		},
	}, nil
}

// EndExecutor handles the "end" node type. It is a no-op that marks workflow completion.
type EndExecutor struct{}

func (e *EndExecutor) Execute(_ context.Context, node Node, _ *ExecutionState) (*StepResult, error) {
	return &StepResult{
		NodeID: node.ID, NodeType: node.Type, Label: node.Data.Label,
		Status: "completed",
		Output: map[string]any{"message": "Workflow execution completed"},
	}, nil
}

// evaluateCondition compares temperature against threshold using the given operator.
// Both values are rounded to 1 decimal place to avoid floating-point precision issues.
func evaluateCondition(temperature float64, operator string, threshold float64) bool {
	t := math.Round(temperature*10) / 10
	th := math.Round(threshold*10) / 10

	switch operator {
	case "greater_than":
		return t > th
	case "less_than":
		return t < th
	case "equals":
		return t == th
	case "greater_than_or_equal":
		return t >= th
	case "less_than_or_equal":
		return t <= th
	default:
		return false
	}
}

func operatorSymbol(op string) string {
	switch op {
	case "greater_than":
		return ">"
	case "less_than":
		return "<"
	case "equals":
		return "="
	case "greater_than_or_equal":
		return ">="
	case "less_than_or_equal":
		return "<="
	default:
		return "?"
	}
}

func operatorLabel(op string) string {
	switch op {
	case "greater_than":
		return "greater than"
	case "less_than":
		return "less than"
	case "equals":
		return "equal to"
	case "greater_than_or_equal":
		return "greater than or equal to"
	case "less_than_or_equal":
		return "less than or equal to"
	default:
		return op
	}
}

// toFloat64 converts an any value to float64, handling json.Number and numeric types.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
