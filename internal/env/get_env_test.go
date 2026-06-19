package env_test

import (
	"cmp"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrelas6/secretas/internal/env"
)

func TestLoadEnv(t *testing.T) {
	pathToEnv := setupEnvfile(t, "")

	_, err := env.NewEnv(pathToEnv)

	if err != nil {
		t.Fatalf("Could not parse env file: %v", err)
	}
}

func TestGetEnv(t *testing.T) {
	pathToEnv := setupEnvfile(t, "")
	env, err := env.NewEnv(pathToEnv)

	if err != nil {
		t.Fatalf("Could not parse env file: %v", err)
	}

	port := env.GetEnv("PORT")
	if port != "3001" {
		t.Fatalf("Expected 3001. Got %s", port)
	}
}

func TestGetEnvFallsBackToEmpty(t *testing.T) {
	pathToEnv := setupEnvfile(t, "")
	env, err := env.NewEnv(pathToEnv)

	if err != nil {
		t.Fatalf("Could not parse env file: %v", err)
	}

	val := env.GetEnv("NOT_IN_ENV_FILE")
	if val != "" {
		t.Fatalf("Expected empty string. Got %s", val)
	}
}

func TestGetEnvFallsBackToOsVar(t *testing.T) {
	pathToEnv := setupEnvfile(t, "")
	t.Setenv("NOT_IN_ENV_FILE", "value")
	env, err := env.NewEnv(pathToEnv)

	if err != nil {
		t.Fatalf("Could not parse env file: %v", err)
	}

	val := env.GetEnv("NOT_IN_ENV_FILE")
	if val != "value" {
		t.Fatalf("Expected value. Got %s", val)
	}
}

func TestGetEnvSkipsCommentsAndBlankLines(t *testing.T) {
	pathToEnv := setupEnvfile(t, "# some comment\n\nPORT=3001")
	env, err := env.NewEnv(pathToEnv)

	if err != nil {
		t.Fatalf("Could not parse env file: %v", err)
	}

	port := env.GetEnv("PORT")

	if port != "3001" {
		t.Fatalf("Expected 3001. Got %s", port)
	}
}

func TestGetEnvDoesNotErrorWhenMissingValue(t *testing.T) {
	t.Setenv("NOT_IN_ENV_FILE", "value")
	env, err := env.NewEnv(".non-existent-env")

	if err != nil {
		t.Fatalf("Could not parse env file: %v", err)
	}

	val := env.GetEnv("NOT_IN_ENV_FILE")
	if val != "value" {
		t.Fatalf("Expected value. Got %s", val)
	}
}

func setupEnvfile(t *testing.T, envFileContent string) string {
	path := filepath.Join(t.TempDir(), ".env")

	content := cmp.Or(envFileContent, "PORT=3001\n")

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("Could not create temp env file: %v", err)
	}

	return path
}
