package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/segersniels/config"
	"github.com/urfave/cli/v2"
)

var AppVersion string
var AppName string

type ConfigData struct {
	LowerCaseFirstLetter     bool `json:"lower_case_first_letter"`
	PromptForOptionalSubType bool `json:"prompt_for_optional_sub_type"`
}

var CONFIG = config.NewConfig("convit", ConfigData{
	LowerCaseFirstLetter:     true,
	PromptForOptionalSubType: true,
})

func main() {
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

							fmt.Println("Configuration updated successfully!")

							return CONFIG.Save()
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
