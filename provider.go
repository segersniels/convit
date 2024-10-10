package main

import (
	"os"

	"github.com/charmbracelet/log"
)

type Provider struct {
	client MessageClient
}

func NewProvider(model string) *Provider {
	var (
		client MessageClient
		apiKey string
	)

	// Depending on the user selected model, we need to set the corresponding API key
	switch model {
	case Claude3Dot5Sonnet:
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			log.Fatal("ANTHROPIC_API_KEY is not set")
		}

		client = NewAnthropic(apiKey, model)
	default:
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Fatal("OPENAI_API_KEY is not set")
		}

		client = NewOpenAI(apiKey, model)
	}

	return &Provider{
		client,
	}
}
