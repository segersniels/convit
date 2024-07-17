package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	openai "github.com/sashabaranov/go-openai"
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
					log.Debug("Ignoring", "file", file)
					shouldIgnore = true
				}
			}

			if !shouldIgnore {
				log.Debug("Using", "header", header)
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
	prompt := fmt.Sprintf("%s\n\n%s", msg, prepareDiff(diff))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: CONFIG.Data.GenerateModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: CONFIG.Data.GenerateSystemMessage,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	log.Debug("System", "message", CONFIG.Data.GenerateSystemMessage)
	log.Debug("Prompt", "prompt", prompt)

	if err != nil {
		return "", err
	}

	log.Debug("Run status", "usage", resp.Usage)

	return resp.Choices[0].Message.Content, nil
}
