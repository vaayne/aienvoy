package bard

// BardAnswer represents the response from the Bard AI service. It contains
// the generated text content, conversation ID, response ID, factuality queries,
// original text query, and any choices provided.
type Answer struct {
	Content           string
	ConversationID    string
	ResponseID        string
	FactualityQueries []any
	TextQuery         string
	Choices           []Choice
	Links             []string
	Images            []any
	ProgramLang       string
	Code              string
	StatusCode        int
}

// Choice represents an alternative response option provided by Bard.
type Choice struct {
	// ID is a unique identifier for the choice.
	ID string
	// Content is the text content of the choice.
	Content string
}
