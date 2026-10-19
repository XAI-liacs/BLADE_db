package components

import "XAI-liacs/BLADE_db/internal/database"

type ExperimentRuns struct {
	Experiment database.Experiment
	Runs       []database.Run
}
