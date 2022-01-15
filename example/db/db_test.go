package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/reiot101/spansqlx"
	"github.com/reiot101/spansqlx/example/db"
)

func NewTestClient(t *testing.T, reset bool) (*spansqlx.DB, error) {
	t.Helper()

	ctx := context.Background()

	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	database := "projects/sandbox/instances/sandbox/databases/sandbox"
	if v := os.Getenv("SPANNER_DATABASE"); v != "" {
		database = v
	}
	// create spanner instance
	if err := db.CreateInstance(ctx, database); err != nil {
		t.Fatal(err)
	}
	// create spanner database
	if err := db.CreateDatabase(ctx, database, reset); err != nil {
		t.Fatal(err)
	}

	path := "file://./../migrations"
	if v := os.Getenv("SPANNER_MIGRATION"); v != "" {
		path = v
	}
	// migrate schema to spanner
	if err := db.Migrate(database, path, db.UP); err != nil {
		t.Fatal(err)
	}

	return spansqlx.Open(context.Background(), spansqlx.WithDatabase(database))
}

func TestPing(t *testing.T) {
	client, err := NewTestClient(t, true)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Ping(context.TODO()); err != nil {
		t.Fatal(err)
	}
}
