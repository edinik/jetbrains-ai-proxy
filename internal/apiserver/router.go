package apiserver

import (
	"fmt"
	"github.com/labstack/echo"
	"jetbrains-ai-proxy/internal/jetbrains"
	"jetbrains-ai-proxy/internal/middleware"
	"jetbrains-ai-proxy/internal/types"
	"jetbrains-ai-proxy/internal/utils"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

func RegisterRoutes(e *echo.Echo) {
	e.Use(middleware.BearerAuth())
	e.POST("/v1/chat/completions", handleChatCompletion)
	e.GET("/v1/models", handleListModels)
}

func handleChatCompletion(c echo.Context) error {
	var req openai.ChatCompletionRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request payload",
		})
	}

	_, err := types.GetModelByName(req.Model)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": fmt.Sprintf("Model '%s' not supported", req.Model),
		})
	}

	if len(req.Messages) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "No messages found",
		})
	}

	jetbrainsReq, err := types.ChatGPTToJetbrainsAI(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
	}

	stream, err := jetbrains.SendJetbrainsRequest(c.Request().Context(), jetbrainsReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer stream.RawBody().Close()

	// 根据请求的 stream 参数决定使用哪种处理方式
	fingerprint := utils.RandStringUsingMathRand(10)
	if req.Stream {
		// 流式处理
		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Transfer-Encoding", "chunked")
		c.Response().WriteHeader(http.StatusOK)

		return jetbrains.StreamJetbrainsAISSEToClient(c.Request().Context(), req, c.Response().Writer, stream.RawBody(), fingerprint)
	} else {
		// 非流式处理
		response, err := jetbrains.ResponseJetbrainsAIToClient(c.Request().Context(), req, stream.RawBody(), fingerprint)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, response)
	}
}

func handleListModels(c echo.Context) error {
	models := types.GetSupportedModels()
	return c.JSON(http.StatusOK, models)
}
