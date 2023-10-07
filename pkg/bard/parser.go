package bard

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
)

func parse(data string) (answer *Answer, err error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("parse bard response error", "err", err)
		}
	}()

	var data1 [][]any
	err = unmarshal([]byte(data), &data1)
	var data2 []any
	err = unmarshal([]byte(data1[0][2].(string)), &data2)

	content := data2[4].([]any)
	links := extractLinks(content)

	// text query
	textQuery := ""
	if data2[2] != nil {
		data3, ok := data2[2].([]any)
		if ok {
			textQuery = data3[0].([]any)[0].(string)
		}
	}

	// program lang
	lang, code := extractProgramLang(content)

	factualityQueries := make([]any, 0)
	if data2[3] != nil {
		factualityQueries = data2[3].([]any)
	}

	choices := extractChoices(content)

	answer = &Answer{
		Content:           choices[0].Content,
		ConversationID:    data2[1].([]any)[0].(string),
		ResponseID:        data2[1].([]any)[1].(string),
		FactualityQueries: factualityQueries,
		Choices:           choices,
		TextQuery:         textQuery,
		Links:             links,
		ProgramLang:       lang,
		Code:              code,
	}
	return
}

// unmarshal json bytes to golang struce, out must be map or point
func unmarshal(in []byte, out any) error {
	err := json.Unmarshal(in, out)
	if err != nil {
		return fmt.Errorf("unmarshal json error: %w", err)
	}
	return nil
}

func extractLinks(data []any) []string {
	links := make([]string, 0)
	for _, item := range data {
		switch reflect.TypeOf(item).Kind() {
		case reflect.Array:
			links = append(links, extractLinks(item.([]any))...)
		case reflect.String:
			s := item.(string)
			if strings.HasPrefix(s, "http") && !strings.Contains(s, "favicon") {
				links = append(links, s)
			}
		}
	}
	return links
}

func extractProgramLang(in []any) (lang, code string) {
	defer func() {
		if r := recover(); r != nil {
			slog.Debug("failed to get program lang")
		}
	}()

	codeBlock := in[4].([]any)[0].([]any)[1].(string)
	codes := strings.Split(codeBlock, "```")
	realCode := codes[1]
	lang = strings.TrimSpace(strings.Split(realCode, "\n")[0])
	code = realCode[len(lang):]
	return
}

func extractChoices(in []any) []Choice {
	choices := make([]Choice, 0)
	for _, item := range in {
		c := item.([]any)
		choices = append(choices, Choice{
			ID:      c[0].(string),
			Content: c[1].([]any)[0].(string),
		})
	}
	return choices
}
