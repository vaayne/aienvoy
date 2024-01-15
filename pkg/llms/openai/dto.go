package openai

import (
	"encoding/json"

	"github.com/Vaayne/aienvoy/pkg/llms/llm"
	"github.com/sashabaranov/go-openai"
)

func toOpenAIChatCompletionRequest(req llm.ChatCompletionRequest) openai.ChatCompletionRequest {
	data, _ := json.Marshal(req)
	var resp openai.ChatCompletionRequest
	_ = json.Unmarshal(data, &resp)
	newModel, ok := modelMappings[resp.Model]
	if ok {
		resp.Model = newModel
	}
	return resp
}

func toLLMChatCompletionResponse(resp openai.ChatCompletionResponse) llm.ChatCompletionResponse {
	data, _ := json.Marshal(resp)
	var req llm.ChatCompletionResponse
	_ = json.Unmarshal(data, &req)
	return req
}

// func toOpenAIChatCompletionStreamResponse(resp llm.ChatCompletionStreamResponse) openai.ChatCompletionStreamResponse {
// 	data, _ := json.Marshal(resp)
// 	var req openai.ChatCompletionStreamResponse
// 	_ = json.Unmarshal(data, &req)
// 	return req
// }

func toLLMChatCompletionStreamResponse(resp openai.ChatCompletionStreamResponse) llm.ChatCompletionStreamResponse {
	data, _ := json.Marshal(resp)
	var req llm.ChatCompletionStreamResponse
	_ = json.Unmarshal(data, &req)
	return req
}
