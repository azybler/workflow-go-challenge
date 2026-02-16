package workflow

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

// HandleGetWorkflow loads a workflow definition from the database and returns it as JSON.
func (s *Service) HandleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	slog.Debug("Getting workflow", "id", id)

	wf, err := s.repo.Get(r.Context(), id)
	if err != nil {
		slog.Error("Failed to get workflow", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if wf == nil {
		writeError(w, http.StatusNotFound, "workflow not found")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wf)
}

// HandleExecuteWorkflow parses execution input, traverses the workflow graph,
// and returns step-by-step results.
func (s *Service) HandleExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	slog.Debug("Executing workflow", "id", id)

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if err := validateExecuteRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	wf, err := s.repo.Get(r.Context(), id)
	if err != nil {
		slog.Error("Failed to get workflow for execution", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if wf == nil {
		writeError(w, http.StatusNotFound, "workflow not found")
		return
	}

	state := &ExecutionState{
		FormData:  req.FormData,
		Condition: req.Condition,
		Variables: make(map[string]any),
	}

	results, err := s.engine.Execute(r.Context(), wf, state)
	if err != nil {
		slog.Error("Workflow execution failed", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

var validOperators = map[string]bool{
	"greater_than":          true,
	"less_than":             true,
	"equals":                true,
	"greater_than_or_equal": true,
	"less_than_or_equal":    true,
}

func validateExecuteRequest(req ExecuteRequest) error {
	if req.FormData == nil {
		return errMissing("formData")
	}
	for _, field := range []string{"name", "email", "city"} {
		v, ok := req.FormData[field]
		if !ok {
			return errMissing(field)
		}
		if s, ok := v.(string); !ok || s == "" {
			return errMissing(field)
		}
	}
	if !validOperators[req.Condition.Operator] {
		return errInvalid("operator")
	}
	return nil
}

type validationError struct {
	field string
	kind  string
}

func (e *validationError) Error() string {
	if e.kind == "missing" {
		return e.field + " is required"
	}
	return e.field + " is invalid"
}

func errMissing(field string) error { return &validationError{field: field, kind: "missing"} }
func errInvalid(field string) error { return &validationError{field: field, kind: "invalid"} }
