package components

import (
	"XAI-liacs/BLADE_db/internal/database"
	"encoding/json"
	"fmt"
	"strings"
)

func renderMarkdown(message string) string {
	return message
	// var buf bytes.Buffer
	// if err := goldmark.Convert([]byte(message), &buf); err != nil {
	// 	return message
	// }
	// return buf.String()
}

func ParseLLM(llm database.Llm) (hardware, config map[string]any) {
	if llm.Hardware.Valid {
		hardware = make(map[string]any)
		if err := json.Unmarshal(
			llm.Hardware.RawMessage,
			&hardware,
		); err != nil {
			hardware = nil
		}
	}
	config = make(map[string]any)
	if err := json.Unmarshal(
		llm.Config,
		&config,
	); err != nil {
		config = nil
	}
	return hardware, config
}

func ParseMethod(llm database.Method) (config map[string]any) {
	config = make(map[string]any)
	if err := json.Unmarshal(
		llm.Config,
		&config,
	); err != nil {
		config = nil
	}
	return config
}

func CleanDictionary(dict map[string]any) map[string]string {
	returndict := make(map[string]string)
	for key, value := range dict {
		key := strings.ReplaceAll(key, "_", " ")
		if value == nil {
			returndict[key] = "-"
		} else {
			returndict[key] = fmt.Sprintf("%v", value)
		}
	}
	return returndict
}
