package types

import (
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

const (
	ChatStreamV7 = "https://api.jetbrains.ai/user/v5/llm/chat/stream/v7"
	PROMPT       = "ij.chat.request.new-chat"
	JwtTokenKey  = "grazie-authenticate-jwt"
)

var modelMap = map[string]OpenAIModel{
	"gpt-4o":      {Object: "model", OwnedBy: "openai", Profile: "openai-gpt-4o"},
	"o1":          {Object: "model", OwnedBy: "openai", Profile: "openai-o1"},
	"o3":          {Object: "model", OwnedBy: "openai", Profile: "openai-o3"},
	"o3-mini":     {Object: "model", OwnedBy: "openai", Profile: "openai-o3-mini"},
	"o4-mini":     {Object: "model", OwnedBy: "openai", Profile: "openai-o4-mini"},
	"gpt4.1":      {Object: "model", OwnedBy: "openai", Profile: "openai-gpt4.1"},
	"gpt4.1-mini": {Object: "model", OwnedBy: "openai", Profile: "openai-gpt4.1-mini"},
	"gpt4.1-nano": {Object: "model", OwnedBy: "openai", Profile: "openai-gpt4.1-nano"},

	"gemini-pro-2.5":   {Object: "model", OwnedBy: "google", Profile: "google-chat-gemini-pro-2.5"},
	"gemini-flash-2.0": {Object: "model", OwnedBy: "google", Profile: "google-chat-gemini-flash-2.0"},
	"gemini-flash-2.5": {Object: "model", OwnedBy: "google", Profile: "google-chat-gemini-flash-2.5"},

	"claude-3.5-haiku":  {Object: "model", OwnedBy: "anthropic", Profile: "anthropic-claude-3.5-haiku"},
	"claude-3.5-sonnet": {Object: "model", OwnedBy: "anthropic", Profile: "anthropic-claude-3.5-sonnet"},
	"claude-3.7-sonnet": {Object: "model", OwnedBy: "anthropic", Profile: "anthropic-claude-3.7-sonnet"},
	"claude-4-sonnet":   {Object: "model", OwnedBy: "anthropic", Profile: "anthropic-claude-4-sonnet"},
}

type OpenAIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
	Profile string `json:"profile"`
}

type OpenAIModelList struct {
	Object string        `json:"object"`
	Data   []OpenAIModel `json:"data"`
}

type MessageField struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
}

type JetbrainsRequest struct {
	Prompt  string    `json:"prompt"`
	Profile string    `json:"profile"`
	Chat    ChatField `json:"chat"`
}

type ChatField struct {
	MessageField []MessageField `json:"messages"`
}

func ChatGPTToJetbrainsAI(chatReq openai.ChatCompletionRequest) (*JetbrainsRequest, error) {
	messageFields, err := convertOpenAIMessagesToJetbrains(chatReq.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	openaiModel, err := GetModelByName(chatReq.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	mReq := &JetbrainsRequest{
		Prompt:  PROMPT,
		Profile: openaiModel.Profile,
		Chat: ChatField{
			MessageField: messageFields,
		},
	}
	if jsonData, err := json.MarshalIndent(mReq, "", "  "); err == nil {
		fmt.Printf("mReq JSON: %s\n", string(jsonData))
	}

	return mReq, nil
}

func convertOpenAIMessagesToJetbrains(openaiMessages []openai.ChatCompletionMessage) ([]MessageField, error) {
	var messageField []MessageField

	for _, msg := range openaiMessages {
		if msg.Role == "system" {
			messageField = append(messageField, MessageField{
				Type:    "system_message",
				Content: msg.Content,
			})
		} else if msg.Role == "user" {
			messageField = append(messageField, MessageField{
				Type:    "user_message",
				Content: msg.Content,
			})
		} else if msg.Role == "assistant" {
			messageField = append(messageField, MessageField{
				Type:    "assistant_message",
				Content: msg.Content,
			})
		}
	}
	return messageField, nil
}

func GetModelByName(modelName string) (OpenAIModel, error) {
	model, exists := modelMap[modelName]
	if !exists {
		return OpenAIModel{}, fmt.Errorf("model '%s' not found", modelName)
	}
	return model, nil
}

func GetSupportedModels() OpenAIModelList {
	var modelSlice []OpenAIModel
	for id, model := range modelMap {
		modelWithID := model
		modelWithID.ID = id
		modelSlice = append(modelSlice, modelWithID)
	}

	return OpenAIModelList{
		Object: "list",
		Data:   modelSlice,
	}
}
