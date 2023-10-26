package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Gets the operating system
func GetOS(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/c", "ver")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting os: %w", err)
	}
	osVersion := strings.TrimSpace(string(output))
	if strings.Contains(osVersion, "Windows") {
		osVersion = "Windows"
	}
	return osVersion, nil
}
