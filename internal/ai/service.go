package ai

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/sashabaranov/go-openai"
)

type Service struct {
	client *openai.Client
	model  string
}

func NewService() (*Service, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)

	return &Service{
		client: client,
		model:  model,
	}, nil
}

func (s *Service) ReviewCalendar(ctx context.Context, calendarJSON string) (string, error) {
	prompt, err := s.loadPrompt()
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	tmpl, err := template.New("prompt").Parse(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var promptBuffer bytes.Buffer
	data := struct {
		CalendarJSON string
	}{
		CalendarJSON: calendarJSON,
	}

	if err := tmpl.Execute(&promptBuffer, data); err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: promptBuffer.String(),
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

func (s *Service) loadPrompt() (string, error) {
	promptPath := "prompts/calendar_review.txt"
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", promptPath, err)
	}

	return string(content), nil
}