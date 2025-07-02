package utils

import (
	"fmt"
	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

func CalculateTokens(text string) int {
	encoding := "cl100k_base"
	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		err = fmt.Errorf("getEncoding: %v", err)
		return 0
	}
	token := tke.Encode(text, nil, nil)
	return len(token)
}

func CalculateJetbrainsUsage(completionText string, spent int) openai.Usage {
	completionTokens := CalculateTokens(completionText)
	return openai.Usage{
		PromptTokens:     spent - completionTokens,
		CompletionTokens: spent - completionTokens,
		TotalTokens:      spent,
	}
}
