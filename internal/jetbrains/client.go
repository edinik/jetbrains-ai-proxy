package jetbrains

import (
	"context"
	"github.com/go-resty/resty/v2"
	"jetbrains-ai-proxy/internal/config"
	"jetbrains-ai-proxy/internal/types"
	"jetbrains-ai-proxy/internal/utils"
	"log"
)

func SendJetbrainsRequest(ctx context.Context, req *types.JetbrainsRequest) (*resty.Response, error) {
	resp, err := utils.RestySSEClient.R().
		SetContext(ctx).
		SetHeader(types.JwtTokenKey, config.JetbrainsAiConfig.JetbrainsToken).
		SetDoNotParseResponse(true).
		SetBody(req).
		Post(types.ChatStreamV7)

	if err != nil {
		log.Printf("jetbrains ai req error: %v", err)
		return nil, err
	}

	return resp, nil
}
