package components

import (
	"XAI-liacs/BLADE_db/internal/database"
	"encoding/json"
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
