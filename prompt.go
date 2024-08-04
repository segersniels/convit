package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

const SYSTEM_MESSAGE string = `Generate a conventional commit message that follows the Conventional Commits specification as described below.
A scope may be provided to a commit’s type, to provide additional contextual information and is contained within parenthesis, e.g., feat(parser): add ability to parse arrays.
Base yourself on the adjusted files in the diff and the actual code changes to determine what the type and scope of the message should be.
Don't include a message body, just the commit title. Don't surround it in backticks or anything of custom markdown formatting.`

const (
	FULL_SUFFIX  = "You will be given a diff of the changes made to the codebase. You will need to generate a full commit message that includes the type, optional scope, and description of the changes."
	SHORT_SUFFIX = "It is your job to come up with only the type and optional scope based on the provided commit message and staged changes (diff) and then reply with the full commit message. Don't touch the original provided commit message, just include it and don't add stuff to it."
)

var FILES_TO_IGNORE = []string{
	"package-lock.json",
	"yarn.lock",
	"npm-debug.log",
	"yarn-debug.log",
	"yarn-error.log",
	".pnpm-debug.log",
	"Cargo.lock",
	"Gemfile.lock",
	"mix.lock",
	"Pipfile.lock",
	"composer.lock",
	"go.sum",
}

func splitDiffIntoChunks(diff string) []string {
	split := strings.Split(diff, "diff --git")[1:]
	for i, chunk := range split {
		split[i] = strings.TrimSpace(chunk)
	}

	return split
}

func removeLockFiles(chunks []string) []string {
	var wg sync.WaitGroup

	filtered := make(chan string)

	for _, chunk := range chunks {
		wg.Add(1)

		go func(chunk string) {
			defer wg.Done()
			shouldIgnore := false
			header := strings.Split(chunk, "\n")[0]

			// Check if the first line contains any of the files to ignore
			for _, file := range FILES_TO_IGNORE {
				if strings.Contains(header, file) {
					shouldIgnore = true
				}
			}

			if !shouldIgnore {
				filtered <- chunk
			}
		}(chunk)
	}

	go func() {
		wg.Wait()
		close(filtered)
	}()

	var result []string
	for chunk := range filtered {
		result = append(result, chunk)
	}

	return result
}

// Split the diff in chunks and remove any lock files to save on tokens
func prepareDiff(diff string) string {
	chunks := splitDiffIntoChunks(diff)

	return strings.Join(removeLockFiles(chunks), "\n")
}

func prepareSystemMessage(full bool) string {
	examples := "Example of the types with the description when they should be used:\n"
	for _, ct := range CommitTypes {
		examples += fmt.Sprintf("- %s: %s\n", ct.Type, ct.Description)
	}

	// If a full generation is requested make sure we explicitly mention that we want a full commit message
	if full {
		return fmt.Sprintf("%s\n\n%s\n\n%s", CONFIG.Data.GenerateSystemMessage, examples, FULL_SUFFIX)
	}

	return fmt.Sprintf("%s\n\n%s\n\n%s", CONFIG.Data.GenerateSystemMessage, examples, SHORT_SUFFIX)
}

func getStagedChanges() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	stdout, err := cmd.Output()

	if err != nil {
		return "", err
	}

	if len(stdout) == 0 {
		return "", errors.New("no staged changes found")
	}

	return string(stdout), nil
}
