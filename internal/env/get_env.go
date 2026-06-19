package env

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

func NewEnv(pathToEnvFile string) (*Env, error) {

	envVars, err := loadEnvFile(pathToEnvFile)

	if err != nil {
		return nil, err
	}

	return &Env{
		vars: envVars,
	}, nil
}

type Env struct {
	vars [][2]string
}

func (e *Env) GetEnv(name string) string {
	for _, row := range e.vars {
		key := row[0]

		if key == name {
			return row[1]
		}
	}

	// fallback
	systemEnvVar, _ := os.LookupEnv(name)
	return systemEnvVar
}

func loadEnvFile(envFilePath string) ([][2]string, error) {
	var envVar [2]string
	var envVars [][2]string

	file, err := os.Open(envFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return envVars, nil
	}

	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		line_trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(line_trimmed, "#") {
			continue
		}

		key, value, found := strings.Cut(line_trimmed, "=")
		if !found {
			continue
		} else {
			envVar[0] = key
			envVar[1] = value
			envVars = append(envVars, envVar)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return envVars, nil
}
