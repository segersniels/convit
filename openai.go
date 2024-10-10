package main

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

var _ MessageClient = (*OpenAI)(nil)

type OpenAI struct {
	apiKey string
	model  string
}

func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{
		apiKey,
		model,
	}
}

func (o *OpenAI) CreateMessage(ctx context.Context, system string, prompt string) (string, error) {
	client := openai.NewClient(o.apiKey)
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: o.model,
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

	return resp.Choices[0].Message.Content, nil
}
