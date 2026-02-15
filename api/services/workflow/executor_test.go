package workflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWeatherClient implements WeatherClient for testing.
type mockWeatherClient struct {
	temperature float64
	err         error
}

func (m *mockWeatherClient) GetTemperature(_ context.Context, _, _ float64) (float64, error) {
	return m.temperature, m.err
}

func newTestState() *ExecutionState {
	return &ExecutionState{
		FormData: map[string]any{
			"name":  "Alice",
			"email": "alice@example.com",
			"city":  "Sydney",
		},
		Condition: ConditionInput{Operator: "greater_than", Threshold: 25},
		Variables: map[string]any{},
	}
}

func integrationNode() Node {
	return Node{
		ID: "weather-api", Type: "integration",
		Data: NodeData{
			Label: "Weather API", Description: "Fetch temperature",
			Metadata: map[string]any{
				"apiEndpoint": "https://api.open-meteo.com/v1/forecast",
				"options": []any{
					map[string]any{"city": "Sydney", "lat": -33.8688, "lon": 151.2093},
					map[string]any{"city": "Melbourne", "lat": -37.8136, "lon": 144.9631},
				},
			},
		},
	}
}

func TestStartExecutor(t *testing.T) {
	exec := &StartExecutor{}
	node := Node{ID: "start", Type: "start", Data: NodeData{Label: "Start"}}

	result, err := exec.Execute(context.Background(), node, newTestState())

	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.NotEmpty(t, result.Output["message"])
}

func TestFormExecutor_Success(t *testing.T) {
	exec := &FormExecutor{}
	node := Node{ID: "form", Type: "form", Data: NodeData{Label: "User Input"}}

	result, err := exec.Execute(context.Background(), node, newTestState())

	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.NotNil(t, result.Output["formData"])
	assert.Contains(t, result.Output["message"].(string), "Alice")
}

func TestFormExecutor_MissingField(t *testing.T) {
	tests := []struct {
		name    string
		remove  string
		wantErr string
	}{
		{"missing name", "name", "name"},
		{"missing email", "email", "email"},
		{"missing city", "city", "city"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := &FormExecutor{}
			state := newTestState()
			delete(state.FormData, tt.remove)

			_, err := exec.Execute(context.Background(), Node{}, state)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestFormExecutor_EmptyField(t *testing.T) {
	exec := &FormExecutor{}
	state := newTestState()
	state.FormData["name"] = ""

	_, err := exec.Execute(context.Background(), Node{}, state)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestIntegrationExecutor_Success(t *testing.T) {
	client := &mockWeatherClient{temperature: 28.5}
	exec := &IntegrationExecutor{client: client}
	state := newTestState()

	result, err := exec.Execute(context.Background(), integrationNode(), state)

	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, 28.5, state.Variables["temperature"])
	assert.Contains(t, result.Output["message"].(string), "28.5")
	assert.Contains(t, result.Output["message"].(string), "Sydney")
}

func TestIntegrationExecutor_CityNotFound(t *testing.T) {
	client := &mockWeatherClient{temperature: 20}
	exec := &IntegrationExecutor{client: client}
	state := newTestState()
	state.FormData["city"] = "Tokyo"

	_, err := exec.Execute(context.Background(), integrationNode(), state)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Tokyo")
	assert.Contains(t, err.Error(), "not found")
}

func TestIntegrationExecutor_APIError(t *testing.T) {
	client := &mockWeatherClient{err: fmt.Errorf("connection timeout")}
	exec := &IntegrationExecutor{client: client}

	_, err := exec.Execute(context.Background(), integrationNode(), newTestState())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "weather API error")
}

func TestConditionExecutor_AllOperators(t *testing.T) {
	tests := []struct {
		operator    string
		temperature float64
		threshold   float64
		want        bool
	}{
		{"greater_than", 28.5, 25, true},
		{"greater_than", 20.0, 25, false},
		{"greater_than", 25.0, 25, false},
		{"less_than", 20.0, 25, true},
		{"less_than", 28.5, 25, false},
		{"less_than", 25.0, 25, false},
		{"equals", 25.0, 25, true},
		{"equals", 25.1, 25.1, true},
		{"equals", 25.0, 25.1, false},
		{"greater_than_or_equal", 25.0, 25, true},
		{"greater_than_or_equal", 28.5, 25, true},
		{"greater_than_or_equal", 20.0, 25, false},
		{"less_than_or_equal", 25.0, 25, true},
		{"less_than_or_equal", 20.0, 25, true},
		{"less_than_or_equal", 28.5, 25, false},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%.1f_%s_%.1f", tt.temperature, tt.operator, tt.threshold)
		t.Run(name, func(t *testing.T) {
			exec := &ConditionExecutor{}
			node := Node{ID: "cond", Type: "condition", Data: NodeData{Label: "Check"}}
			state := &ExecutionState{
				Condition: ConditionInput{Operator: tt.operator, Threshold: tt.threshold},
				Variables: map[string]any{"temperature": tt.temperature},
			}

			result, err := exec.Execute(context.Background(), node, state)

			require.NoError(t, err)
			assert.Equal(t, "completed", result.Status)
			assert.Equal(t, tt.want, result.Output["conditionMet"])

			expectedHandle := "false"
			if tt.want {
				expectedHandle = "true"
			}
			assert.Equal(t, expectedHandle, state.Variables["conditionResult"])
		})
	}
}

func TestConditionExecutor_MissingTemperature(t *testing.T) {
	exec := &ConditionExecutor{}
	state := &ExecutionState{
		Condition: ConditionInput{Operator: "greater_than", Threshold: 25},
		Variables: map[string]any{},
	}

	_, err := exec.Execute(context.Background(), Node{}, state)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "temperature")
}

func TestEmailExecutor(t *testing.T) {
	exec := &EmailExecutor{}
	node := Node{
		ID: "email", Type: "email",
		Data: NodeData{
			Label: "Send Alert",
			Metadata: map[string]any{
				"emailTemplate": map[string]any{
					"subject": "Weather Alert",
					"body":    "Weather alert for {{city}}! Temperature is {{temperature}}\u00b0C!",
				},
			},
		},
	}
	state := &ExecutionState{
		FormData:  map[string]any{"name": "Alice", "email": "alice@example.com", "city": "Sydney"},
		Variables: map[string]any{"temperature": 28.5},
	}

	result, err := exec.Execute(context.Background(), node, state)

	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Contains(t, result.Output["message"].(string), "alice@example.com")

	draft := result.Output["emailDraft"].(map[string]any)
	assert.Equal(t, "alice@example.com", draft["to"])
	assert.Contains(t, draft["body"].(string), "Sydney")
	assert.Contains(t, draft["body"].(string), "28.5")
}

func TestEndExecutor(t *testing.T) {
	exec := &EndExecutor{}
	node := Node{ID: "end", Type: "end", Data: NodeData{Label: "Complete"}}

	result, err := exec.Execute(context.Background(), node, newTestState())

	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.NotEmpty(t, result.Output["message"])
}

func TestEvaluateCondition_FloatRounding(t *testing.T) {
	// 0.1 + 0.2 should equal 0.3 after rounding
	assert.True(t, evaluateCondition(0.1+0.2, "equals", 0.3))
}
