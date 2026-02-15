package workflow

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkflowRepo abstracts workflow persistence for testability.
type WorkflowRepo interface {
	Get(ctx context.Context, id string) (*Workflow, error)
}

// Service wires together the repository and execution engine for the workflow domain.
type Service struct {
	repo   WorkflowRepo
	engine *Engine
}

// NewService creates a Service with a real PostgreSQL repository and Open-Meteo weather client.
func NewService(pool *pgxpool.Pool) (*Service, error) {
	repo := NewRepository(pool)
	weatherClient := NewOpenMeteoClient()
	registry := NewRegistry(weatherClient)
	engine := NewEngine(registry)
	return &Service{repo: repo, engine: engine}, nil
}

// jsonMiddleware sets the Content-Type header to application/json.
func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// LoadRoutes registers workflow HTTP handlers on the given router.
func (s *Service) LoadRoutes(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix("/workflows").Subrouter()
	router.StrictSlash(false)
	router.Use(jsonMiddleware)

	router.HandleFunc("/{id}", s.HandleGetWorkflow).Methods("GET")
	router.HandleFunc("/{id}/execute", s.HandleExecuteWorkflow).Methods("POST")
}
