package openai

import (
	"github.com/Vaayne/aienvoy/pkg/llms/llm"
	"github.com/mitchellh/mapstructure"
	"github.com/sashabaranov/go-openai"
)

func toOpenAIChatCompletionRequest(req llm.ChatCompletionRequest) openai.ChatCompletionRequest {
	req.Model = req.ModelId()
	var resp openai.ChatCompletionRequest
	_ = mapstructure.Decode(req, &resp)
	return resp
}

func toLLMChatCompletionStreamResponse(resp openai.ChatCompletionStreamResponse) llm.ChatCompletionStreamResponse {
	var llmResp llm.ChatCompletionStreamResponse
	_ = mapstructure.Decode(resp, &llmResp)
	return llmResp
}
