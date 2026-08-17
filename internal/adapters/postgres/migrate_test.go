package postgres

import (
	"io/fs"
	"strings"
	"testing"
)

// A migration file that no code path applies is invisible until a query needs
// its columns in production, so the registry is checked against the directory
// rather than against a list somebody has to remember to extend.
func TestEveryEmbeddedFileIsRegisteredInOrder(t *testing.T) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no migrations are embedded")
	}
	ordered, err := migrations()
	if err != nil {
		t.Fatal(err)
	}
	if len(ordered) != len(entries) {
		t.Fatalf("registered %d migrations for %d files", len(ordered), len(entries))
	}
	for i, entry := range entries {
		version, name, err := migrationName(entry.Name())
		if err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		if ordered[i].version != version || ordered[i].name != name {
			t.Errorf("file %s registered as %03d %s", entry.Name(), ordered[i].version, ordered[i].name)
		}
		if strings.TrimSpace(ordered[i].up) == "" || strings.TrimSpace(ordered[i].down) == "" {
			t.Errorf("file %s has an empty up or down section", entry.Name())
		}
	}
}

func TestMigrationNameRejectsUnusableFiles(t *testing.T) {
	for _, name := range []string{"run_profiles.sql", "6_run_profiles.sql", "0006_run_profiles.sql", "006.sql", "abc_run_profiles.sql", "000_zero.sql"} {
		if _, _, err := migrationName(name); err == nil {
			t.Errorf("accepted %q", name)
		}
	}
	version, name, err := migrationName("006_run_profiles.sql")
	if err != nil || version != 6 || name != "run_profiles" {
		t.Fatalf("got %d %q %v", version, name, err)
	}
}

func TestMigrationPartsRequiresBothSections(t *testing.T) {
	if _, _, err := migrationParts("-- migrate:up\nCREATE TABLE t();\n"); err == nil {
		t.Error("accepted a migration with no down section")
	}
	up, down, err := migrationParts("-- migrate:up\nCREATE TABLE t();\n-- migrate:down\nDROP TABLE t;\n")
	if err != nil || !strings.Contains(up, "CREATE TABLE") || !strings.Contains(down, "DROP TABLE") {
		t.Fatalf("got up=%q down=%q err=%v", up, down, err)
	}
}
