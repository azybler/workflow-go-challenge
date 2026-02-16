package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubRepo implements the Get method for testing without a database.
type stubRepo struct {
	workflow *Workflow
	err      error
}

func (r *stubRepo) Get(_ context.Context, _ string) (*Workflow, error) {
	return r.workflow, r.err
}

func newTestService(wf *Workflow, weatherTemp float64) *Service {
	repo := &stubRepo{workflow: wf}
	client := &mockWeatherClient{temperature: weatherTemp}
	registry := NewRegistry(client)
	engine := NewEngine(registry)
	return &Service{repo: repo, engine: engine}
}

func setupRouter(svc *Service) *mux.Router {
	router := mux.NewRouter()
	sub := router.PathPrefix("/api/v1/workflows").Subrouter()
	sub.HandleFunc("/{id}", svc.HandleGetWorkflow).Methods("GET")
	sub.HandleFunc("/{id}/execute", svc.HandleExecuteWorkflow).Methods("POST")
	return router
}

func TestHandleGetWorkflow_Success(t *testing.T) {
	wf := testWorkflow()
	svc := newTestService(wf, 0)
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/api/v1/workflows/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result Workflow
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "test-wf", result.ID)
	assert.Len(t, result.Nodes, 6)
	assert.Len(t, result.Edges, 6)
}

func TestHandleGetWorkflow_NotFound(t *testing.T) {
	svc := newTestService(nil, 0)
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/api/v1/workflows/00000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	assert.Equal(t, "workflow not found", result["message"])
}

func TestHandleGetWorkflow_InvalidID(t *testing.T) {
	svc := newTestService(nil, 0)
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/api/v1/workflows/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	assert.Equal(t, "invalid workflow id", result["message"])
}

func TestHandleExecuteWorkflow_Success(t *testing.T) {
	wf := testWorkflow()
	svc := newTestService(wf, 30.0)
	router := setupRouter(svc)

	body, _ := json.Marshal(ExecuteRequest{
		FormData:  map[string]any{"name": "Alice", "email": "alice@example.com", "city": "Sydney"},
		Condition: ConditionInput{Operator: "greater_than", Threshold: 25},
	})

	req := httptest.NewRequest("POST", "/api/v1/workflows/550e8400-e29b-41d4-a716-446655440000/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result ExecutionResults
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.NotEmpty(t, result.Steps)
	assert.NotEmpty(t, result.ExecutionID)
}

func TestHandleExecuteWorkflow_BadInput(t *testing.T) {
	wf := testWorkflow()
	svc := newTestService(wf, 30.0)
	router := setupRouter(svc)

	body, _ := json.Marshal(ExecuteRequest{
		FormData:  map[string]any{"name": "Alice"},
		Condition: ConditionInput{Operator: "greater_than", Threshold: 25},
	})

	req := httptest.NewRequest("POST", "/api/v1/workflows/550e8400-e29b-41d4-a716-446655440000/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	assert.Contains(t, result["message"], "required")
}

func TestHandleExecuteWorkflow_InvalidOperator(t *testing.T) {
	wf := testWorkflow()
	svc := newTestService(wf, 30.0)
	router := setupRouter(svc)

	body, _ := json.Marshal(ExecuteRequest{
		FormData:  map[string]any{"name": "Alice", "email": "alice@example.com", "city": "Sydney"},
		Condition: ConditionInput{Operator: "invalid_op", Threshold: 25},
	})

	req := httptest.NewRequest("POST", "/api/v1/workflows/550e8400-e29b-41d4-a716-446655440000/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	assert.Contains(t, result["message"], "operator")
}

func TestHandleExecuteWorkflow_NotFound(t *testing.T) {
	svc := newTestService(nil, 0)
	router := setupRouter(svc)

	body, _ := json.Marshal(ExecuteRequest{
		FormData:  map[string]any{"name": "Alice", "email": "alice@example.com", "city": "Sydney"},
		Condition: ConditionInput{Operator: "greater_than", Threshold: 25},
	})

	req := httptest.NewRequest("POST", "/api/v1/workflows/00000000-0000-0000-0000-000000000000/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleExecuteWorkflow_InvalidJSON(t *testing.T) {
	wf := testWorkflow()
	svc := newTestService(wf, 30.0)
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/api/v1/workflows/550e8400-e29b-41d4-a716-446655440000/execute", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
