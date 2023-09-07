package bard

// BardAnswer represents the response from the Bard AI service. It contains
// the generated text content, conversation ID, response ID, factuality queries,
// original text query, and any choices provided.
type BardAnswer struct {
	// Content is the generated text content.
	Content string
	// ConversationID is the unique ID for the conversation with Bard.
	ConversationID string
	// ResponseID is the unique ID for this particular response.
	ResponseID string
	// FactualityQueries are any follow up factuality clarification queries.
	FactualityQueries []any
	// TextQuery is the original text query sent to Bard.
	TextQuery string
	// Choices are any alternative responses provided.
	Choices []Choice
}

// Choice represents an alternative response option provided by Bard.
type Choice struct {
	// ID is a unique identifier for the choice.
	ID string
	// Content is the text content of the choice.
	Content string
}
