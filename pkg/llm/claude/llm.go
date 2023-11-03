package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llm"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

const (
	ModelClaudeV2            = "anthropic.claude-v2"
	ModelClaudeV1Dot3        = "anthropic.claude-v1"
	ModelClaudeInstantV1Dot2 = "anthropic.claude-instant-v1"
)

type Claude struct {
	*Client
	llm.LLM
}

func New(config aws.Config) *Claude {
	return &Claude{
		Client: NewClient(config),
		LLM:    llm.LLM{},
	}
}

func (c *Claude) ListModels() []string {
	return ListModels()
}

func ListModels() []string {
	return []string{ModelClaudeV2, ModelClaudeV1Dot3, ModelClaudeInstantV1Dot2}
}

func (c *Claude) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	slog.InfoContext(ctx, "chat start", "model", req.Model, "is_stream", false)
	bedrockRequest := &BedrockRequest{}
	bedrockRequest.FromChatCompletionRequest(req)

	output, err := c.Client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(req.Model),
		Body:        bedrockRequest.Marshal(),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		slog.ErrorContext(ctx, "chat start", "model", req.Model, "is_stream", false, "err", err)
		return llm.ChatCompletionResponse{}, err
	}
	resp := &BedrockResponse{}
	resp.Unmarshal(output.Body)
	slog.InfoContext(ctx, "chat success", "model", req.Model, "is_stream", false)
	return resp.ToChatCompletionResponse(), nil
}

func (c *Claude) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	slog.InfoContext(ctx, "chat start", "model", req.Model, "is_stream", true)
	bedrockRequest := &BedrockRequest{}
	bedrockRequest.FromChatCompletionRequest(req)

	output, err := c.Client.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(req.Model),
		Body:        bedrockRequest.Marshal(),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		errChan <- err
		return
	}

	sb := &strings.Builder{}

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var resp BedrockResponse
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				slog.ErrorContext(ctx, "chat start", "model", req.Model, "is_stream", true, "err", err)
				errChan <- err
				return
			}
			sb.WriteString(resp.Completion)
			dataChan <- resp.ToChatCompletionStreamResponse()
		case *types.UnknownUnionMember:
			err = fmt.Errorf("unknown event type: %T", v)
			slog.ErrorContext(ctx, "chat start", "model", req.Model, "is_stream", true, "err", err)
			errChan <- err
			return
		default:
			err = fmt.Errorf("unknown event type: %T", v)
			slog.ErrorContext(ctx, "chat start", "model", req.Model, "is_stream", true, "err", err)
			errChan <- err
			return
		}
	}
	slog.InfoContext(ctx, "chat success", "model", req.Model, "is_stream", true)
	errChan <- io.EOF
}
