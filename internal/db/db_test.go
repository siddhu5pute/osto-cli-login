package db

import "testing"

func TestRunMigrations_NoErrorOnEmptyDB(t *testing.T) {
	t.Skip("requires live Postgres — wire up in CI with a postgres service container")
}
