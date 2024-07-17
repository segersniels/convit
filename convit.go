package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Convit struct {
	client *OpenAI
}

func NewConvit() *Convit {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if len(apiKey) == 0 {
		return &Convit{
			client: nil,
		}
	}

	return &Convit{
		client: NewOpenAI(apiKey),
	}
}

// Prompt the user for the main commit type and optional sub-type
func (c *Convit) promptForScope() (string, error) {
	var main, opt string
	options := []huh.Option[string]{
		huh.NewOption("feat: Adds or removes a new feature", "feat"),
		huh.NewOption("fix: Fixes a bug", "fix"),
		huh.NewOption("refactor: A code change that neither fixes a bug nor adds a feature, eg. renaming a variable, removing dead code, etc.", "refactor"),
		huh.NewOption("docs: Documentation only changes", "docs"),
		huh.NewOption("style: Changes the style of the code eg. linting", "style"),
		huh.NewOption("perf: Improves the performance of the code", "perf"),
		huh.NewOption("test: Adding missing tests or correcting existing tests", "test"),
		huh.NewOption("chore: Changes that don't change source code or tests", "chore"),
		huh.NewOption("build: Changes that affect the build system or external dependencies (example scopes: gulp, broccoli, npm)", "build"),
		huh.NewOption("ci: Changes to CI configuration files and scripts", "ci"),
		huh.NewOption("revert: Reverts a previous commit", "revert"),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("Select the type of commit").Options(options...).Value(&main).Filtering(true).Validate(func(val string) error {
				if len(val) == 0 {
					return errors.New("type cannot be empty")
				}

				return nil
			}),
		),
		huh.NewGroup(
			huh.NewInput().Title("Provide an optional scope (leave empty for none)").Value(&opt),
		).WithHideFunc(func() bool {
			return !CONFIG.Data.PromptForOptionalSubType
		}),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	// Just use the main type if no optional sub-type is provided
	if len(opt) == 0 {
		return main, nil
	}

	return fmt.Sprintf("%s(%s)", main, opt), nil
}

// Prompts the user for the commit message
func (c *Convit) promptForMessage() (string, error) {
	var msg string
	if err := huh.NewInput().Title("Enter your commit message").Value(&msg).Run(); err != nil {
		return "", err
	}

	if len(msg) == 0 {
		log.Error("Message cannot be empty")
		os.Exit(0)
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

func (c *Convit) Generate() error {
	// Get the commit message
	msg, err := c.promptForMessage()
	if err != nil {
		return err
	}

	// Check if the OpenAI client is initialized
	if c.client == nil {
		return errors.New("\"OPENAI_API_KEY\" is not set")
	}

	diff, err := getStagedChanges()
	if err != nil {
		return err
	}

	var response string
	for {
		if err := spinner.New().TitleStyle(lipgloss.NewStyle()).Title("Generating your commit message...").Action(func() {
			response, err = c.client.GetChatCompletion(diff, msg)
			if err != nil {
				log.Fatal(err)
			}
		}).Run(); err != nil {
			return err
		}

		var confirmation bool
		if err := huh.NewConfirm().Title(response).Description("Do you want to commit this message?").Value(&confirmation).Run(); err != nil {
			return err
		}

		if confirmation {
			break
		}
	}

	cmd := exec.Command("git", "commit", "-m", response)

	return cmd.Run()
}
