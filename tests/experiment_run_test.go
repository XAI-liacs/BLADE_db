package tests

import (
	"XAI-liacs/BLADE_db/internal/config"
	"XAI-liacs/BLADE_db/internal/database"
	"context"
	"testing"

	"github.com/google/uuid"
)

type ExperimentRun struct {
	experiment database.CreateExperimentParams
	run        database.CreateRunParams
}

func prepareExperimentandRuns(state config.State) (experimentRuns []ExperimentRun, err error) {
	preTestClear(state)
	experimentRuns = make([]ExperimentRun, 0)
	experiments := make([]database.CreateExperimentParams, 0)
	runs := make([]database.CreateRunParams, 0)
	for i := range 3 {
		experiment := getRandomExperiment()
		experiments = append(experiments, experiment)

		run := database.CreateRunParams{ID: uuid.New(), Seed: int32(i + 1)}
		runs = append(runs, run)

		_, err = state.DB.CreateExperiment(context.Background(), experiment)
		if err != nil {
			return
		}
		_, err = state.DB.CreateRun(context.Background(), run)
		if err != nil {
			return
		}
	}
	for _, run := range runs {
		for _, experiment := range experiments {
			experimentRuns = append(experimentRuns, ExperimentRun{experiment: experiment, run: run})
		}
	}
	return
}

func TestRelationTableExperimentRunInsertion(t *testing.T) {
	state := config.GetContext("test")
	testCases, err := prepareExperimentandRuns(state)
	if err != nil {
		t.Fatalf("Cannot prepare experiment and run tables: %s", err.Error())
	}

	// Insert all parent-child relations.
	for _, testCase := range testCases {
		err = state.DB.CreateExperimentRun(
			context.Background(),
			database.CreateExperimentRunParams{
				ExperimentID: testCase.experiment.ID,
				RunID:        testCase.run.ID,
			})
		if err != nil {
			t.Fatalf(
				"Unable to insert into experiment_run: %s",
				err,
			)
		}
	}

	// Check whether experiment run are valid.
	for _, testCase := range testCases {
		runs, err := state.DB.GetRunsForExperiment(
			context.Background(),
			testCase.experiment.ID,
		)
		if err != nil {
			t.Fatalf(
				"Unable to get experiments for run %s: %s",
				testCase.experiment.ID,
				err,
			)
		}

		found := false
		for _, run := range runs {
			if run.ID == testCase.run.ID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf(
				"Expected experiment %s to have run %s",
				testCase.experiment.ID.String(),
				testCase.run.ID.String(),
			)
		}
	}
}

func TestRelationTableExperimentRunFollowsForeignKeyPolicy(t *testing.T) {
	state := config.GetContext("test")
	err := state.DB.CreateExperimentRun(context.Background(), database.CreateExperimentRunParams{
		ExperimentID: uuid.New(),
		RunID:        uuid.New(),
	})
	if err == nil {
		t.Fatal("Table `experiment_run`, doesn't follow foreign-key policy.")
	}
}
