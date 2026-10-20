package components

import (
	"XAI-liacs/BLADE_db/internal/database"
)

type ExperimentRuns struct {
	Experiment database.Experiment
	Runs       []database.Run
}

func renderMarkdown(message string) string {
	return message
	// var buf bytes.Buffer
	// if err := goldmark.Convert([]byte(message), &buf); err != nil {
	// 	return message
	// }
	// return buf.String()
}
