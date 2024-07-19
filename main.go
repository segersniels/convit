package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
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
	GenerateModel:            GPT4oMini,
	GenerateSystemMessage: `Generate a conventional commit message that follows the Conventional Commits specification as described below.

A scope may be provided to a commit’s type, to provide additional contextual information and is contained within parenthesis, e.g., feat(parser): add ability to parse arrays.
It is your job to come up with only the type and optional scope based on the provided commit message and staged changes (diff) and then reply with the full commit message.
Don't touch the original provided commit message, just include it and don't add stuff to it.

Base yourself on the adjusted files in the diff and the actual code changes to determine what the type and scope of the message should be.
Don't include a message body, just the commit title. Don't surround it in backticks or anything of custom markdown formatting.`,
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
									models := huh.NewOptions(GPT4oMini, GPT4o, GPT4Turbo, GPT3Dot5Turbo)
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
