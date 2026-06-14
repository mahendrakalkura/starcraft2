package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func sample(ctx context.Context, file string) error {
	if file == "" {
		return fmt.Errorf("--file is required for the sample action")
	}

	sidecar, err := locateSidecar()
	if err != nil {
		return fmt.Errorf("locateSidecar(): %w", err)
	}

	command := exec.CommandContext(ctx, sidecar, file)
	command.Stderr = os.Stderr

	stdout, err := command.Output()
	if err != nil {
		return fmt.Errorf("command.Output(): %w", err)
	}

	pretty := bytes.Buffer{}
	err = json.Indent(&pretty, stdout, "", "    ")
	if err != nil {
		return fmt.Errorf("json.Indent(): %w", err)
	}

	fmt.Println(pretty.String())

	return nil
}
