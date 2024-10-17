package main

import (
	"context"
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

const (
	GPT4o             = "gpt-4o"
	GPT4oMini         = "gpt-4o-mini"
	GPT4Turbo         = "gpt-4-turbo"
	GPT3Dot5Turbo     = "gpt-3.5-turbo"
	Claude3Dot5Sonnet = "claude-3-5-sonnet-20240620"
)

const (
	MessageRoleSystem    = "system"
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
)

type MessageClient interface {
	CreateMessage(ctx context.Context, system string, prompt string) (string, error)
}

type ConfigData struct {
	LowerCaseFirstLetter     bool   `json:"lower_case_first_letter"`
	PromptForOptionalSubType bool   `json:"prompt_for_optional_sub_type"`
	GenerateModel            string `json:"generate_model"`
	GenerateSystemMessage    string `json:"generate_prompt"`
}

var CONFIG = config.NewConfig("convit", ConfigData{
	LowerCaseFirstLetter:     true,
	PromptForOptionalSubType: false,
	GenerateModel:            GPT4oMini,
	GenerateSystemMessage:    SYSTEM_MESSAGE,
})

func main() {
	err := checkIfNewVersionIsAvailable()
	if err != nil {
		log.Debug("Failed to check for latest release", "error", err)
	}

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
				Name:  "update",
				Usage: "Update convit to the latest version",
				Action: func(ctx *cli.Context) error {
					return convit.Update()
				},
			},
			{
				Name:  "commit",
				Usage: "Write a commit message",
				Action: func(ctx *cli.Context) error {
					return convit.Commit()
				},
			},
			{
				Name:  "generate",
				Usage: "Write a commit message with the help of AI",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "partial",
						Usage: "Only generate the commit type and scope",
					},
				},
				Action: func(ctx *cli.Context) error {
					return convit.Generate(ctx.Bool("partial"))
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
									models := huh.NewOptions(GPT4oMini, GPT4o, GPT4Turbo, GPT3Dot5Turbo, Claude3Dot5Sonnet)
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
