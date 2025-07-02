package jetbrains

import (
	"bufio"
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/sashabaranov/go-openai"
	"io"
	"jetbrains-ai-proxy/internal/utils"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	sseObject         = "chat.completion.chunk"
	completionsObject = "chat.completions"
	sseFinish         = "[DONE]"
	initialBufferSize = 4096
	maxBufferSize     = 1024 * 1024 // 1MB
	flushThreshold    = 10
	heartbeatInterval = 30 * time.Second
)

type SSEData struct {
	Type      string       `json:"type"`
	EventType string       `json:"event_type"`
	Content   string       `json:"content,omitempty"`
	Reason    string       `json:"reason,omitempty"`
	Updated   *UpdatedData `json:"updated,omitempty"`
	Spent     *SpentData   `json:"spent,omitempty"`
}

type UpdatedData struct {
	License string     `json:"license"`
	Current AmountData `json:"current"`
	Maximum AmountData `json:"maximum"`
	Until   int64      `json:"until"`
	QuotaID QuotaInfo  `json:"quotaID"`
}

type AmountData struct {
	Amount string `json:"amount"`
}

type QuotaInfo struct {
	QuotaId string `json:"quotaId"`
}

type SpentData struct {
	Amount string `json:"amount"`
}

// ResponseJetbrainsAIToClient 处理非流式响应
func ResponseJetbrainsAIToClient(ctx context.Context, req openai.ChatCompletionRequest, r io.Reader, fp string) (openai.ChatCompletionResponse, error) {
	reader := bufio.NewReader(r)
	var fullContent strings.Builder

	now := time.Now().Unix()
	chatId := strconv.Itoa(int(now))

	for {
		select {
		case <-ctx.Done():
			return openai.ChatCompletionResponse{}, ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Reached EOF for non-streaming response")
				break
			}
			return openai.ChatCompletionResponse{}, fmt.Errorf("读取错误: %w", err)
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		jsonStr := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if jsonStr == "" || jsonStr == sseFinish || jsonStr == "end" {
			continue
		}

		var sseData SSEData
		if err := sonic.UnmarshalString(jsonStr, &sseData); err != nil {
			log.Printf("解析SSE数据错误: %v", err)
			continue
		}

		if sseData.Type == "Content" {
			fullContent.WriteString(sseData.Content)
		}

		if sseData.Type == "QuotaMetadata" {
			var spentAmount float64
			if sseData.Spent != nil {
				if amount, err := strconv.ParseFloat(sseData.Spent.Amount, 64); err == nil {
					spentAmount = amount
				} else {
					log.Printf("Warning: failed to parse spent amount '%s': %v", sseData.Spent.Amount, err)
				}
			}
			usage := utils.CalculateJetbrainsUsage(fullContent.String(), int(math.Round(spentAmount)))
			return createMessage(chatId, now, req, usage, fullContent.String(), fp), nil
		}
	}

	// 如果没有收到 QuotaMetadata，返回默认响应
	usage := utils.CalculateJetbrainsUsage(fullContent.String(), 0)
	return createMessage(chatId, now, req, usage, fullContent.String(), fp), nil
}

// StreamJetbrainsAISSEToClient 处理流式响应
func StreamJetbrainsAISSEToClient(ctx context.Context, req openai.ChatCompletionRequest, w io.Writer, r io.Reader, fp string) error {
	log.Printf("=== Starting SSE Stream Processing for model: %s ===", req.Model)

	reader := bufio.NewReaderSize(r, initialBufferSize)
	writer := bufio.NewWriterSize(w, initialBufferSize)

	now := time.Now().Unix()
	chatId := strconv.Itoa(int(now))
	fingerprint := fp

	log.Printf("Session initialized - ChatID: %s, Fingerprint: %s", chatId, fingerprint)

	var completionBuilder strings.Builder
	messageCount := 0
	totalBufferSize := 0

	// 创建心跳检测器
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-heartbeat.C:
			if err := sendHeartbeat(writer, w); err != nil {
				log.Printf("Heartbeat error: %v", err)
			}
			continue
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Reached EOF after %d messages", messageCount)
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		log.Printf("Received line: %s", strings.TrimSpace(line))

		// 检查缓冲区大小
		totalBufferSize += len(line)
		if totalBufferSize > maxBufferSize {
			log.Printf("Buffer overflow: current size %d exceeds max size %d", totalBufferSize, maxBufferSize)
			return fmt.Errorf("buffer overflow: exceeded maximum buffer size of %d bytes", maxBufferSize)
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		jsonStr := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if jsonStr == "" || jsonStr == "end" {
			continue
		}

		var sseData SSEData
		if err := sonic.UnmarshalString(jsonStr, &sseData); err != nil {
			log.Printf("Error unmarshaling SSE data: %v", err)
			continue
		}

		log.Printf("Received SSE data: %+v", sseData)

		messageCount++

		if err := processMessage(writer, w, sseData, chatId, fingerprint, now, &completionBuilder, req); err != nil {
			log.Printf("Failed to process message: %v", err)
			return err
		}

		// 定期刷新缓冲区
		if messageCount >= flushThreshold {
			if err := flushWriter(writer, w); err != nil {
				return fmt.Errorf("flush error: %w", err)
			}
			messageCount = 0
		}

		// 检查是否结束
		if sseData.Type == "QuotaMetadata" {
			if err := sendFinishSignal(writer, w); err != nil {
				return fmt.Errorf("finish signal error: %w", err)
			}
			log.Printf("Stream completed successfully")
			return nil
		}
	}
}

// processMessage 处理单个消息
func processMessage(writer *bufio.Writer, w io.Writer, sseData SSEData, chatId, fingerprint string, now int64, completionBuilder *strings.Builder, req openai.ChatCompletionRequest) error {
	switch sseData.Type {
	case "Content":
		completionBuilder.WriteString(sseData.Content)
		sseMsg := createStreamMessage(chatId, now, req, fingerprint, sseData.Content, "")
		return sendMessage(writer, w, sseMsg)

	case "QuotaMetadata":
		var spentAmount float64
		if sseData.Spent != nil {
			if amount, err := strconv.ParseFloat(sseData.Spent.Amount, 64); err == nil {
				spentAmount = amount
			} else {
				log.Printf("Warning: failed to parse spent amount '%s': %v", sseData.Spent.Amount, err)
			}
		}

		usage := utils.CalculateJetbrainsUsage(completionBuilder.String(), int(math.Round(spentAmount)))
		sseMsg := createStreamMessage(chatId, now, req, fingerprint, "", "")
		sseMsg.Choices[0].FinishReason = openai.FinishReasonStop
		sseMsg.Usage = &usage
		return sendMessage(writer, w, sseMsg)

	default:
		// 忽略其他类型的消息
		log.Printf("Ignoring message type: %s", sseData.Type)
		return nil
	}
}

// createStreamMessage 创建流式消息
func createStreamMessage(chatId string, now int64, req openai.ChatCompletionRequest, fingerPrint string, content string, reasoningContent string) openai.ChatCompletionStreamResponse {
	choice := openai.ChatCompletionStreamChoice{
		Index: 0,
		Delta: openai.ChatCompletionStreamChoiceDelta{
			Role:             openai.ChatMessageRoleAssistant,
			Content:          content,
			ReasoningContent: reasoningContent,
		},
		ContentFilterResults: openai.ContentFilterResults{},
		FinishReason:         openai.FinishReasonNull,
	}

	return openai.ChatCompletionStreamResponse{
		ID:                "chatcmpl-" + chatId,
		Object:            sseObject,
		Created:           now,
		Model:             req.Model,
		Choices:           []openai.ChatCompletionStreamChoice{choice},
		SystemFingerprint: fingerPrint,
	}
}

// createMessage 创建非流式消息响应
func createMessage(chatId string, now int64, req openai.ChatCompletionRequest, usage openai.Usage, content string, fp string) openai.ChatCompletionResponse {
	choice := openai.ChatCompletionChoice{
		Index: 0,
		Message: openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		},
		FinishReason: openai.FinishReasonStop,
	}

	return openai.ChatCompletionResponse{
		ID:                "chatcmpl-" + chatId,
		Object:            completionsObject,
		Created:           now,
		Model:             req.Model,
		Choices:           []openai.ChatCompletionChoice{choice},
		SystemFingerprint: fp,
		Usage:             usage,
	}
}

// sendMessage 发送消息到客户端
func sendMessage(writer *bufio.Writer, w io.Writer, sseMsg openai.ChatCompletionStreamResponse) error {
	sendLine, err := sonic.MarshalString(sseMsg)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	outputMsg := fmt.Sprintf("data: %s\n\n", sendLine)
	if _, err := writer.WriteString(outputMsg); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	return flushWriter(writer, w)
}

// sendHeartbeat 发送心跳包
func sendHeartbeat(writer *bufio.Writer, w io.Writer) error {
	if _, err := writer.WriteString(": keepalive\n\n"); err != nil {
		return fmt.Errorf("heartbeat write error: %w", err)
	}
	return flushWriter(writer, w)
}

// sendFinishSignal 发送结束信号
func sendFinishSignal(writer *bufio.Writer, w io.Writer) error {
	finishMsg := fmt.Sprintf("data: %s\n\n", sseFinish)
	if _, err := writer.WriteString(finishMsg); err != nil {
		return fmt.Errorf("write finish signal error: %w", err)
	}
	return flushWriter(writer, w)
}

// flushWriter 刷新写入器
func flushWriter(writer *bufio.Writer, w io.Writer) error {
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}
