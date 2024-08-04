package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type CommitType struct {
	Type        string
	Description string
	SubType     string
}

var CommitTypes = []CommitType{
	{Type: "chore", Description: "Changes that don't change source code or tests"},
	{Type: "feat", Description: "Adds or removes a new feature"},
	{Type: "fix", Description: "Fixes a bug"},
	{Type: "refactor", Description: "A code change that neither fixes a bug nor adds a feature, eg. renaming a variable, remove dead code, etc."},
	{Type: "docs", Description: "Documentation only changes"},
	{Type: "style", Description: "Changes the style of the code eg. linting"},
	{Type: "perf", Description: "Improves the performance of the code"},
	{Type: "test", Description: "Adding missing tests or correcting existing tests"},
	{Type: "build", Description: "Changes that affect the build system or external dependencies (example scopes: gulp, broccoli, npm)"},
	{Type: "ci", Description: "Changes to CI configuration files and scripts"},
	{Type: "revert", Description: "Reverts a previous commit"},
	{Type: "chore", SubType: "release", Description: "Release / Version tags"},
	{Type: "chore", SubType: "deps", Description: "Add, remove or update dependencies"},
	{Type: "chore", SubType: "dev-deps", Description: "Add, remove or update development dependencies"},
}

type Convit struct {
	client MessageClient
}

func NewConvit() *Convit {
	var (
		client MessageClient
		apiKey string
	)

	// Depending on the user selected model, we need to set the corresponding API key
	switch CONFIG.Data.GenerateModel {
	case Claude3Dot5Sonnet:
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			log.Fatal("ANTHROPIC_API_KEY is not set")
		}

		client = NewAnthropic(apiKey, CONFIG.Data.GenerateModel)
	default:
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Fatal("OPENAI_API_KEY is not set")
		}

		client = NewOpenAI(apiKey, CONFIG.Data.GenerateModel)
	}

	return &Convit{
		client,
	}
}

// promptForScope prompts the user for the main commit type and optional sub-type
func (c *Convit) promptForScope() (string, error) {
	var main, opt string

	options := make([]huh.Option[string], 0, len(CommitTypes))
	for _, ct := range CommitTypes {
		optionText := fmt.Sprintf("%s: %s", ct.Type, ct.Description)
		optionValue := ct.Type

		// If there's a sub-type associated with the commit type, include it in the option text and value
		if ct.SubType != "" {
			optionValue = fmt.Sprintf("%s(%s)", ct.Type, ct.SubType)
			optionText = fmt.Sprintf("%s: %s", optionValue, ct.Description)
		}

		options = append(options, huh.NewOption(optionText, optionValue))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select the type of commit").
				Options(options...).
				Value(&main).
				Filtering(true).
				Validate(func(val string) error {
					if val == "" {
						return errors.New("type cannot be empty")
					}

					return nil
				}),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Provide an optional scope (leave empty for none)").
				Value(&opt),
		).WithHideFunc(func() bool {
			// If the user selects a type with a sub-type, we don't need to ask for the sub-type
			if regexp.MustCompile(`\((.*?)\)`).MatchString(main) {
				return true
			}

			// If the user doesn't want to be prompted for an optional sub-type, skip the sub-type prompt
			if !CONFIG.Data.PromptForOptionalSubType {
				return true
			}

			return false
		}),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	// If the user didn't provide an optional sub-type, just return the main type
	if opt == "" {
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

func (c *Convit) Generate(partial bool) error {
	var msg *string
	if partial {
		message, err := c.promptForMessage()
		if err != nil {
			return err
		}

		msg = &message
	}

	diff, err := getStagedChanges()
	if err != nil {
		return err
	}

	var response string
	for {
		if err := spinner.New().TitleStyle(lipgloss.NewStyle()).Title("Generating your commit message...").Action(func() {
			system := prepareSystemMessage(partial)
			diff := prepareDiff(diff)

			// If partial generation is requested, we need to add the user specified message to the prompt
			if partial {
				diff = fmt.Sprintf("message: %s\n\ndiff: %s", *msg, diff)
			}

			// Set a timeout for the request
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			response, err = c.client.CreateMessage(ctx, system, diff)
			if err != nil {
				log.Fatal(err)
			}
		}).Run(); err != nil {
			return err
		}

		// If the response is empty don't bother asking the user for confirmation
		if len(response) == 0 {
			return errors.New("failed to generate commit message")
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
