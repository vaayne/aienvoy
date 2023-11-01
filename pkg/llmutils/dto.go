package llmutils

import "github.com/sashabaranov/go-openai"

func CompletionRequestToChatCompletionRequest(req openai.CompletionRequest) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: req.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: req.Prompt.(string),
			},
		},
		MaxTokens:        req.MaxTokens,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		N:                req.N,
		Stream:           req.Stream,
		Stop:             req.Stop,
		PresencePenalty:  req.PresencePenalty,
		FrequencyPenalty: req.FrequencyPenalty,
		LogitBias:        req.LogitBias,
		User:             req.User,
	}
}

func ChatCompletionResponseToCompletionResponse(resp openai.ChatCompletionResponse) openai.CompletionResponse {
	choices := make([]openai.CompletionChoice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = openai.CompletionChoice{
			Text:         choice.Message.Content,
			Index:        choice.Index,
			FinishReason: string(choice.FinishReason),
		}
	}

	return openai.CompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Usage:   resp.Usage,
		Choices: choices,
	}
}

func ChatCompletionStreamResponseToCompletionResponse(resp openai.ChatCompletionStreamResponse) openai.CompletionResponse {
	choices := make([]openai.CompletionChoice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = openai.CompletionChoice{
			Text:         choice.Delta.Content,
			Index:        choice.Index,
			FinishReason: string(choice.FinishReason),
		}
	}

	return openai.CompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
	}
}
