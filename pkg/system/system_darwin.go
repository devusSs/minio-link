package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Gets the operating system
func GetOS(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "sw_vers", "-productName")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting os: %w", err)
	}
	osName := strings.TrimSpace(string(output))
	return osName, nil
}
