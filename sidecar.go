package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func locateSidecar() (string, error) {
	location := os.Getenv("SC2JSON")
	if location != "" {
		return location, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("os.Executable(): %w", err)
	}

	location = filepath.Join(filepath.Dir(executable), "sc2json")
	_, err = os.Stat(location)
	if err != nil {
		return "", fmt.Errorf("sc2json not found at %s (set SC2JSON to override): %w", location, err)
	}

	return location, nil
}

func runSidecar(ctx context.Context, sidecar string, path string) (*Replay, error) {
	command := exec.CommandContext(ctx, sidecar, path)

	stderr := bytes.Buffer{}
	command.Stderr = &stderr

	stdout, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}

	replay := &Replay{}
	err = json.Unmarshal(stdout, replay)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal(): %w", err)
	}

	return replay, nil
}
