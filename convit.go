package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

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
	{Type: "chore", Description: "Changes that don't change source code or tests"},
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

// Prompt the user for the main commit type and optional sub-type
func (c *Convit) promptForScope() (string, error) {
	var main, opt string
	options := make([]huh.Option[string], len(CommitTypes))
	for i, ct := range CommitTypes {
		if ct.SubType != "" {
			options[i] = huh.NewOption(fmt.Sprintf("%s(%s): %s", ct.Type, ct.SubType, ct.Description), fmt.Sprintf("%s(%s)", ct.Type, ct.SubType))
		} else {
			options[i] = huh.NewOption(fmt.Sprintf("%s: %s", ct.Type, ct.Description), ct.Type)
		}
	}

	var shouldAskForSubType = true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("Select the type of commit").Options(options...).Value(&main).Filtering(true).Validate(func(val string) error {
				if len(val) == 0 {
					return errors.New("type cannot be empty")
				}

				// There's already an optional subtype associated with the chosen option, so we don't need to show the subtype prompt
				if regexp.MustCompile(`\((.*?)\)`).MatchString(val) {
					shouldAskForSubType = false
				}

				return nil
			}),
		),
		huh.NewGroup(
			huh.NewInput().Title("Provide an optional scope (leave empty for none)").Value(&opt),
		).WithHideFunc(func() bool {
			return !CONFIG.Data.PromptForOptionalSubType || !shouldAskForSubType
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
			response, err = c.client.CreateMessage(diff, msg)
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
