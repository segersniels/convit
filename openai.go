package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	openai "github.com/sashabaranov/go-openai"
)

const (
	GPT4oMini     = "gpt-4o-mini"
	GPT4o         = "gpt-4o"
	GPT4Turbo     = "gpt-4-turbo"
	GPT3Dot5Turbo = "gpt-3.5-turbo"
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
	filtered := make(chan string, len(chunks))

	for _, chunk := range chunks {
		wg.Add(1)
		go func(chunk string) {
			defer wg.Done()
			header := strings.Split(chunk, "\n")[0]

			for _, file := range FILES_TO_IGNORE {
				if strings.Contains(header, file) {
					log.Debug("Ignoring", "file", file)
					return
				}
			}

			log.Debug("Using", "header", header)
			filtered <- chunk
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

func prepareSystemMessage() string {
	examples := "Example of the types with the description when they should be used:\n"
	for _, ct := range CommitTypes {
		examples += fmt.Sprintf("- %s: %s\n", ct.Type, ct.Description)
	}

	return fmt.Sprintf("%s\n\n%s", CONFIG.Data.GenerateSystemMessage, examples)
}

type OpenAI struct {
	ApiKey string
}

func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{
		ApiKey: apiKey,
	}
}

func (o *OpenAI) GetChatCompletion(diff string, msg string) (string, error) {
	client := openai.NewClient(o.ApiKey)
	system := prepareSystemMessage()
	prompt := fmt.Sprintf("message: %s\n\ndiff: %s", msg, prepareDiff(diff))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: CONFIG.Data.GenerateModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: system,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	log.Debug("", "system", system)
	log.Debug("", "prompt", prompt)
	log.Debug("", "usage", resp.Usage)

	return resp.Choices[0].Message.Content, nil
}
