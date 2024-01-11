package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	configfileBs, err := os.ReadFile(home + "/.explain.json")
	if err != nil {
		panic(err)
	}

	var configFile struct {
		OPENAI_API_KEY string
	}

	err = json.Unmarshal(configfileBs, &configFile)
	if err != nil {
		panic(err)
	}

	promptParts := os.Args[1:]
	prompt := strings.Join(promptParts, " ")

	client := openai.NewClient(configFile.OPENAI_API_KEY)
	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4TurboPreview,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: `You are a helpful terminal assistant.`,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   1500,
		Temperature: 0,
		Stream:      true,
	})
	if err != nil {
		panic(err)
	}

	for {
		res, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			panic(err)
		}

		fmt.Print(res.Choices[0].Delta.Content)
	}
	fmt.Println()
}
