package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

var (
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorBlue  = "\033[34m"
	colorReset = "\033[0m"
)

var (
	flagClearConversation = flag.Bool("clear", false, "Clear the conversation history")
	flagChangeModel       = flag.String("model", "", "Change the model used for the conversation")
	flagInit              = flag.Bool("init", false, "Initialize the configuration file")
	flagConfig            = flag.Bool("config", false, "Show the current configuration")
	flagConversation      = flag.Bool("conversation", false, "Show the current conversation")
)

var configFilePath string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	configFilePath = home + "/.explain.json"
}

func main() {
	flag.Parse()

	var config explainConfig
	if f := *flagInit; f {
		config.setDefault()
		err := config.save()
		if err != nil {
			panic(err)
		}

		fmt.Println("Saved configuration file to", configFilePath)
		yNo := ""
		fmt.Print("Do you want to add an api key? (y/n, default y): ")
		fmt.Scanln(&yNo)

		if yNo == "n" {
			return
		}

		fmt.Println("Please provide your OpenAI API key")
		fmt.Print("API key: ")
		fmt.Scanln(&config.OpenaiApiKey)
		err = config.save()
		if err != nil {
			panic(err)
		}

		fmt.Println()

		fmt.Println("All good, you can start using explain now!\nExample usage: `$ explain what is the meaning of life`")
		return
	}

	err := config.readFromFile()
	if err != nil {
		fmt.Println("Failed to read the configuration file")
		fmt.Println("Please run `explain -init` to create a new configuration file")
		return
	}

	if f := *flagClearConversation; f {
		err := config.clearConversation()
		if err != nil {
			panic(err)
		}
		return
	}

	if f := *flagChangeModel; f != "" {
		model, ok := parseModel(f)
		if !ok {
			fmt.Printf(
				"Invalid model %v, please provide one of the following:\n%v\n",
				f, prettyModels(models),
			)
			return
		}

		config.Model = model
		err := config.save()
		if err != nil {
			panic(err)
		}

		return
	}

	if f := flagConfig; *f {
		configfileBs, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(configfileBs))
		return
	}

	if f := flagConversation; *f {
		for _, msg := range config.Conversation {
			switch msg.Role {
			case "user":
				fmt.Print(colorBlue)
			case "assistant":
				fmt.Print(colorGreen)
			case "system":
				fmt.Print(colorRed)
			}

			fmt.Printf("[%v]", msg.Role)
			fmt.Print(colorReset)

			fmt.Printf("\n%v\n", msg.Content)

			fmt.Println()
		}

		return
	}

	if len(config.Conversation) == 0 {
		config.Conversation = []Msg{
			{
				Role:    "system",
				Content: `You will be straight to the point and very concise.`,
			},
		}
		fmt.Println("system: You will be straight to the point and very concise.")
	}

	{ // read the user input
		promptParts := os.Args[1:]
		prompt := strings.Join(promptParts, " ")

		prompt = strings.TrimSpace(prompt)
		if prompt == "" {
			fmt.Println("Please provide a prompt")
			fmt.Println()
			flag.Usage()
			return
		}

		config.Conversation = append(config.Conversation, Msg{
			Role:    "user",
			Content: prompt,
		})
	}

	client := openai.NewClient(config.OpenaiApiKey)
	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		Model:       openai.GPT4TurboPreview,
		Messages:    config.Conversation,
		MaxTokens:   1500,
		Temperature: 0,
		Stream:      true,
	})
	if err != nil {
		panic(err)
	}

	var newMessage string
	for {
		res, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			panic(err)
		}

		messageChunk := res.Choices[0].Delta.Content
		newMessage += messageChunk
		fmt.Print(messageChunk)
	}
	fmt.Println()

	config.Conversation = append(config.Conversation, Msg{
		Role:    "assistant",
		Content: newMessage,
	})

	err = config.save()
	if err != nil {
		panic(err)
	}
}

type explainConfig struct {
	OpenaiApiKey string    `json:"openai_api_key"`
	Model        string    `json:"model"`
	UpdatedAt    time.Time `json:"updated_at"`
	Conversation []Msg     `json:"conversation"`
}

func (this *explainConfig) readFromFile() (err error) {
	defer wrapErr(&err)

	configfileBs, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(configfileBs, this)
	if err != nil {
		return err
	}

	if this.UpdatedAt.IsZero() {
		this.UpdatedAt = time.Now()
	}

	return nil
}

func (this explainConfig) shouldRenewMessages() bool {
	return this.UpdatedAt.Add(24 * time.Hour).Before(time.Now())
}

func (this *explainConfig) clearConversation() (err error) {
	defer wrapErr(&err)

	this.UpdatedAt = time.Now()
	this.Conversation = []Msg{}

	return this.save()
}

func (this *explainConfig) setDefault() {
	*this = explainConfig{
		OpenaiApiKey: "",
		Model:        "gpt-4",
		UpdatedAt:    time.Now(),
		Conversation: []openai.ChatCompletionMessage{},
	}
}

func (this explainConfig) save() (err error) {
	defer wrapErr(&err)

	configfileBs, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFilePath, configfileBs, 0644)
}

func (this explainConfig) addMessages(newMessages []Msg) (err error) {
	defer wrapErr(&err)

	this.Conversation = append(this.Conversation, newMessages...)

	return this.save()
}

var models = []string{
	openai.GPT432K0613,
	openai.GPT432K0314,
	openai.GPT432K,
	openai.GPT40613,
	openai.GPT40314,
	openai.GPT4TurboPreview,
	openai.GPT4VisionPreview,
	openai.GPT4,
	openai.GPT3Dot5Turbo1106,
	openai.GPT3Dot5Turbo0613,
	openai.GPT3Dot5Turbo0301,
	openai.GPT3Dot5Turbo16K,
	openai.GPT3Dot5Turbo16K0613,
	openai.GPT3Dot5Turbo,
	openai.GPT3Dot5TurboInstruct,
}

func parseModel(model string) (string, bool) {
	for _, m := range models {
		if m == model {
			return model, true
		}
	}

	return "", false
}

type prettyModels []string

func (prettyModels) String() string {
	return "  - " + strings.Join(models, "\n  - ")
}
