package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
)

type Convit struct{}

func NewConvit() *Convit {
	return &Convit{}
}

// Prompt the user for an optional sub-type (scope) for the commit
func (c *Convit) promptForOptionalSubType() (string, error) {
	var wantsScope bool
	if err := huh.NewConfirm().Title("Do you want to specify an optional scope?").Value(&wantsScope).Run(); err != nil {
		return "", err
	}

	if !wantsScope {
		return "", nil
	}

	var scope string
	if err := huh.NewInput().Title("Select the optional scope of your commit").Value(&scope).Run(); err != nil {
		return "", err
	}

	return scope, nil
}

// Prompt the user for the main commit type and optional sub-type
func (c *Convit) promptForScope() (string, error) {
	var main string
	options := []huh.Option[string]{
		huh.NewOption("feat: Adds or removes a new feature", "feat"),
		huh.NewOption("fix: Fixes a bug", "fix"),
		huh.NewOption("refactor: Changes the structure of the code", "refactor"),
		huh.NewOption("docs: Updates to documentation", "docs"),
		huh.NewOption("style: Changes the style of the code eg. linting", "style"),
		huh.NewOption("perf: Improves the performance of the code", "perf"),
		huh.NewOption("test: Changes to the tests", "test"),
		huh.NewOption("chore: Miscellaneous commits e.g. modifying .gitignore", "chore"),
		huh.NewOption("build: Changes that affect build components like build tool, ci pipeline, dependencies, project version, etc.", "build"),
		huh.NewOption("ci: Changes to CI configuration files and scripts", "ci"),
		huh.NewOption("revert: Reverts a previous commit", "revert"),
	}

	if err := huh.NewSelect[string]().Title("Select the type of commit").Options(options...).Value(&main).Filtering(true).Run(); err != nil {
		return "", err
	}

	// Ensure the type is not empty
	if len(main) == 0 {
		return "", fmt.Errorf("type cannot be empty")
	}

	// If the user has disabled the prompt for optional sub-types, return the main type
	if !CONFIG.Data.PromptForOptionalSubType {
		return main, nil
	}

	scope := main
	opt, err := c.promptForOptionalSubType()
	if err != nil {
		return "", err
	}

	// Combine main type and optional sub-type if provided
	if len(opt) > 0 {
		scope = fmt.Sprintf("%s(%s)", scope, opt)
	}

	return scope, nil
}

// Prompts the user for the commit message
func (c *Convit) promptForMessage() (string, error) {
	var msg string
	if err := huh.NewInput().Title("Enter your commit message").Value(&msg).Run(); err != nil {
		return "", err
	}

	if len(msg) == 0 {
		return "", fmt.Errorf("message cannot be empty")
	}

	// Ensure the first letter of the message is lowercase
	if CONFIG.Data.LowerCaseFirstLetter && len(msg) > 0 {
		msg = strings.ToLower(msg[:1]) + msg[1:]
	}

	return msg, nil
}

// Prompt user for commit type, scope, and message, then execute the commit
func (c *Convit) Commit() error {
	// Get the commit scope (type and optional sub-type)
	scope, err := c.promptForScope()
	if err != nil {
		return err
	}

	// Get the commit message
	msg, err := c.promptForMessage()
	if err != nil {
		return err
	}

	// Combine scope and message into a conventional commit format
	conv := fmt.Sprintf("%s: %s", scope, msg)

	// Execute the git commit command
	cmd := exec.Command("git", "commit", "-m", conv)

	return cmd.Run()
}
