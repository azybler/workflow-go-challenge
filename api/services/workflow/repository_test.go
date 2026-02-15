package workflow

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping repository tests")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })
	return pool
}

func TestRepository_InitSchema(t *testing.T) {
	pool := getTestPool(t)
	repo := NewRepository(pool)

	err := repo.InitSchema(context.Background())
	require.NoError(t, err)

	// Running again should be idempotent
	err = repo.InitSchema(context.Background())
	require.NoError(t, err)
}

func TestRepository_Seed(t *testing.T) {
	pool := getTestPool(t)
	repo := NewRepository(pool)

	ctx := context.Background()
	require.NoError(t, repo.InitSchema(ctx))

	err := repo.Seed(ctx)
	require.NoError(t, err)
}

func TestRepository_Seed_Idempotent(t *testing.T) {
	pool := getTestPool(t)
	repo := NewRepository(pool)

	ctx := context.Background()
	require.NoError(t, repo.InitSchema(ctx))

	require.NoError(t, repo.Seed(ctx))
	require.NoError(t, repo.Seed(ctx)) // Second call should not error
}

func TestRepository_Get_Found(t *testing.T) {
	pool := getTestPool(t)
	repo := NewRepository(pool)

	ctx := context.Background()
	require.NoError(t, repo.InitSchema(ctx))
	require.NoError(t, repo.Seed(ctx))

	wf, err := repo.Get(ctx, sampleWorkflowID)
	require.NoError(t, err)
	require.NotNil(t, wf)

	assert.Equal(t, sampleWorkflowID, wf.ID)
	assert.Equal(t, "Weather Alert Workflow", wf.Name)
	assert.Len(t, wf.Nodes, 6)
	assert.Len(t, wf.Edges, 6)

	// Verify start node exists
	var hasStart bool
	for _, n := range wf.Nodes {
		if n.Type == "start" {
			hasStart = true
			break
		}
	}
	assert.True(t, hasStart, "workflow should have a start node")
}

func TestRepository_Get_NotFound(t *testing.T) {
	pool := getTestPool(t)
	repo := NewRepository(pool)

	ctx := context.Background()
	require.NoError(t, repo.InitSchema(ctx))

	wf, err := repo.Get(ctx, "00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	assert.Nil(t, wf)
}
