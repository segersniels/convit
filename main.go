package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/sashabaranov/go-openai"
	"github.com/segersniels/config"
	"github.com/urfave/cli/v2"
)

var AppVersion string
var AppName string

type ConfigData struct {
	LowerCaseFirstLetter     bool   `json:"lower_case_first_letter"`
	PromptForOptionalSubType bool   `json:"prompt_for_optional_sub_type"`
	GenerateModel            string `json:"generate_model"`
	GenerateSystemMessage    string `json:"generate_prompt"`
}

var CONFIG = config.NewConfig("convit", ConfigData{
	LowerCaseFirstLetter:     true,
	PromptForOptionalSubType: true,
	GenerateModel:            openai.GPT4o,
	GenerateSystemMessage: `Generate a conventional commit message that follows the Conventional Commits specification as described below.

The commit contains the following structural elements, to communicate intent to the consumers of your library:
1. fix: a commit of the type fix patches a bug in your codebase (this correlates with PATCH in Semantic Versioning).
2. feat: a commit of the type feat introduces a new feature to the codebase (this correlates with MINOR in Semantic Versioning).
3. BREAKING CHANGE: a commit that has a footer BREAKING CHANGE:, or appends a ! after the type/scope, introduces a breaking API change (correlating with MAJOR in Semantic Versioning). A BREAKING CHANGE can be part of commits of any type.
4. types other than fix: and feat: are allowed, for example @commitlint/config-conventional (based on the Angular convention) recommends build:, chore:, ci:, docs:, style:, refactor:, perf:, test:, and others.
5. footers other than BREAKING CHANGE: <description> may be provided and follow a convention similar to git trailer format.

A scope may be provided to a commit’s type, to provide additional contextual information and is contained within parenthesis, e.g., feat(parser): add ability to parse arrays.
It is your job to come up with only the type and optional scope based on the provided commit message and staged changes (diff) and then reply with the full commit message.
Don't touch the original provided commit message, just include it and don't add stuff to it.

Base yourself on the adjusted files in the diff and the actual code changes to determine what the type and scope of the message should be.
Don't include a message body, just the commit title. Don't surround it in backticks or anything of custom markdown formatting.

Example of the types with the description when they should be used:
- feat: Adds or removes a new feature
- fix: Fixes a bug
- refactor: A code change that neither fixes a bug nor adds a feature, eg. renaming a variable, removing dead code, etc.
- docs: Documentation only changes
- style: Changes the style of the code eg. linting
- perf: Improves the performance of the code
- test: Adding missing tests or correcting existing tests
- chore: Changes that don't change source code or tests
- build: Changes that affect the build system or external dependencies (example scopes: gulp, broccoli, npm)
- ci: Changes to CI configuration files and scripts
- revert: Reverts a previous commit.`,
})

func main() {
	debug := os.Getenv("DEBUG")
	if debug != "" {
		log.SetLevel(log.DebugLevel)
	}

	convit := NewConvit()

	app := &cli.App{
		Name:    AppName,
		Usage:   "Write conventional commit messages",
		Version: AppVersion,
		Commands: []*cli.Command{
			{
				Name:  "commit",
				Usage: "Write a commit message",
				Action: func(ctx *cli.Context) error {
					err := convit.Commit()
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "generate",
				Usage: "Write a commit message with the help of OpenAI",
				Action: func(ctx *cli.Context) error {
					err := convit.Generate()
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "config",
				Usage: "Configure the app",
				Subcommands: []*cli.Command{
					{
						Name:  "init",
						Usage: "Initialize the config",
						Action: func(ctx *cli.Context) error {
							form := huh.NewForm(
								huh.NewGroup(
									huh.NewConfirm().
										Title("Lowercase first letter of commit message?").
										Description("This will automatically lowercase the first letter of your commit message.").
										Value(&CONFIG.Data.LowerCaseFirstLetter),
								),
								huh.NewGroup(
									huh.NewConfirm().
										Title("Prompt for optional sub-type?").
										Description("This will ask if you want to specify an optional scope for your commit.").
										Value(&CONFIG.Data.PromptForOptionalSubType),
								),
							)

							err := form.Run()
							if err != nil {
								return err
							}

							log.Info("Configuration updated successfully!")

							return CONFIG.Save()
						},
						Subcommands: []*cli.Command{
							{
								Name:  "ai",
								Usage: "Initialize the AI config",
								Action: func(ctx *cli.Context) error {
									models := huh.NewOptions(openai.GPT4o, openai.GPT4Turbo, openai.GPT3Dot5Turbo)
									form := huh.NewForm(
										huh.NewGroup(
											huh.NewSelect[string]().Title("Model").Description("Configure the default model").Options(models...).Value(&CONFIG.Data.GenerateModel),
											huh.NewText().Title("System Message").Description("Configure the default system message").CharLimit(99999).Value(&CONFIG.Data.GenerateSystemMessage),
										),
									)

									err := form.Run()
									if err != nil {
										return err
									}

									log.Info("Configuration updated successfully!")

									return CONFIG.Save()
								},
							},
						},
					},
					{
						Name:  "ls",
						Usage: "List the current configuration",
						Action: func(ctx *cli.Context) error {
							data, err := json.MarshalIndent(CONFIG.Data, "", "  ")
							if err != nil {
								return err
							}

							fmt.Println(string(data))
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
