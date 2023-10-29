package llmclaude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/llm/usage"

	"github.com/sashabaranov/go-openai"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

const (
	ModelClaudeV2            = "anthropic.claude-v2"
	ModelClaudeV1Dot3        = "anthropic.claude-v1"
	ModelClaudeInstantV1Dot2 = "anthropic.claude-instant-v1"
)

type Claude struct {
	client *bedrockruntime.Client
}

func New() *Claude {
	return &Claude{
		client: bedrockruntime.NewFromConfig(*getAWSConfig()),
	}
}

func getAWSConfig() *aws.Config {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(config.GetConfig().AWS.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.GetConfig().AWS.AccessKeyId,
			config.GetConfig().AWS.SecretAccessKey,
			"",
		)))
	if err != nil {
		slog.Error("get aws config error", "err", err)
		return nil
	}
	return &cfg
}

func (c *Claude) CreateChatCompletion(ctx context.Context, req *openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	bedrockRequest := &BedrockRequest{}
	bedrockRequest.FromOpenAIChatCompletionRequest(req)

	output, err := c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(req.Model),
		Body:        bedrockRequest.Marshal(),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}
	resp := &BedrockResponse{}
	resp.Unmarshal(output.Body)
	return resp.ToOpenAIChatCompletionResponse(), nil
}

func (c *Claude) CreateChatCompletionStream(ctx context.Context, req *openai.ChatCompletionRequest, dataChan chan openai.ChatCompletionStreamResponse, errChan chan error) {
	bedrockRequest := &BedrockRequest{}
	bedrockRequest.FromOpenAIChatCompletionRequest(req)

	output, err := c.client.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
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
				errChan <- err
				return
			}
			sb.WriteString(resp.Completion)
			dataChan <- resp.ToOpenAIChatCompletionStreamResponse()
		case *types.UnknownUnionMember:
			errChan <- fmt.Errorf("unknown union member: %s", v.Tag)
			return
		default:
			errChan <- fmt.Errorf("unknown event type: %T", v)
			return
		}
	}
	_ = usage.SaveFromText(ctx, req.Model, sb.String())
	errChan <- io.EOF
}

func (c *Claude) CreateCompletion(ctx context.Context, req *openai.CompletionRequest) (openai.CompletionResponse, error) {
	bedrockRequest := &BedrockRequest{}
	bedrockRequest.FromOpenAICompletionRequest(req)

	output, err := c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(req.Model),
		Body:        bedrockRequest.Marshal(),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return openai.CompletionResponse{}, err
	}
	resp := &BedrockResponse{}
	resp.Unmarshal(output.Body)
	return resp.ToOpenAICompletionResponse(), nil
}

func (c *Claude) CreateCompletionStream(ctx context.Context, req *openai.CompletionRequest, dataChan chan openai.CompletionResponse, errChan chan error) {
	bedrockRequest := &BedrockRequest{}
	bedrockRequest.FromOpenAICompletionRequest(req)

	output, err := c.client.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
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
				errChan <- err
				return
			}
			sb.WriteString(resp.Completion)
			dataChan <- resp.ToOpenAICompletionResponse()
		case *types.UnknownUnionMember:
			errChan <- fmt.Errorf("unknown union member: %s", v.Tag)
			return
		default:
			errChan <- fmt.Errorf("unknown event type: %T", v)
			return
		}
	}
	_ = usage.SaveFromText(ctx, req.Model, sb.String())
	errChan <- io.EOF
}