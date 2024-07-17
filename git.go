package main

import "os/exec"

func getStagedChanges() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
