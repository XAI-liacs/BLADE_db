package tests

import (
	filehandlers "anantashahane/BLADE_db/file_handlers"
	"anantashahane/BLADE_db/internal/config"
	"context"
	"slices"
	"testing"

	"github.com/google/uuid"
)

func TestOrchestrationImportsExperiment(t *testing.T) {
	state := config.GetContext("test")

	db_scratch_pad := context.Background()

	db_data, err := filehandlers.ImportProgress("./test_files/Erdös_Min_Overlap", state, db_scratch_pad)

	if err != nil {
		t.Fatal("Something went wrong: " + err.Error())
	}

	// Check experiment insertion:
	fetched_experiments, err := state.DB.GetExperiments(context.Background())
	found := false
	for _, experiment := range fetched_experiments {
		if experiment.ID == db_data.Experiment.DatabaseID {
			found = true
		}
	}
	if !found {
		t.Fatal("cannot find entry experiment in db.")
	}
	// Check run + experiment_run insertion:
	fetched_runs, err := state.DB.GetRunsForExperiment(context.Background(), db_data.Experiment.DatabaseID)
	found = false
	for _, run := range db_data.RunData {
		for _, fetched_run := range fetched_runs {
			if fetched_run.ID == run.DatabaseID {
				found = true
			}
		}
		if !found {
			t.Fatal("cannot find entry  in db.")
		}
	}
}

func TestOrchestrationImportsSolution(t *testing.T) {
	state := config.GetContext("test")
	preTestClear(state)
	db_scratch_pad := context.Background()

	db_data, err := filehandlers.ImportProgress("./test_files/Erdös_Min_Overlap", state, db_scratch_pad)
	if err != nil {
		t.Fatal("Something went wrong: " + err.Error())
	}

	file_details := filehandlers.FileDetails{
		Root:   "./test_files/Erdös_Min_Overlap",
		Suffix: "log.jsonl",
	}

	run_solution_ids_mapping := make(map[uuid.UUID][]filehandlers.ID_Mapping)
	for run_dir, run := range db_data.RunData {
		run_file_details := file_details
		run_file_details.Root = run_dir
		solution_ids, err := filehandlers.ImportSolution(run_file_details, run.DatabaseID, state, db_scratch_pad)
		if err != nil {
			t.Fatalf("Unable to import solutions: %s", err.Error())
		}
		run_solution_ids_mapping[run.DatabaseID] = solution_ids
	}

	// Check experiment insertion:
	fetched_experiments, err := state.DB.GetExperiments(context.Background())
	found := false
	for _, experiment := range fetched_experiments {
		if experiment.ID == db_data.Experiment.DatabaseID {
			found = true
		}
	}
	if !found {
		t.Fatal("cannot find entry experiment in db.")
	}
	// Check run + experiment_run insertion:
	fetched_runs, err := state.DB.GetRunsForExperiment(context.Background(), db_data.Experiment.DatabaseID)
	found = false
	for _, run := range db_data.RunData {
		for _, fetched_run := range fetched_runs {
			if fetched_run.ID == run.DatabaseID {
				found = true
			}
		}
		if !found {
			t.Fatal("cannot find entry  in db.")
		}
	}

	// Check solution insertions:
	for _, run := range db_data.RunData {
		db_solutions, err := state.DB.GetSolutionsForRun(db_scratch_pad, run.DatabaseID)
		if err != nil {
			t.Fatal(err)
		}
		if len(db_solutions) != len(run_solution_ids_mapping[run.DatabaseID]) {
			t.Fatalf("Not all solutions imported, for run %v, expected %v, found %v", run.DatabaseID, len(run_solution_ids_mapping[run.DatabaseID]), len(db_solutions))
		}
		found = false
		for _, db_solution := range db_solutions {
			content, err := state.DB.GetChildrenofSolution(db_scratch_pad, db_solution.ID)
			if err != nil {
				return
			}
			if len(content) > 0 {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Run %v, did not model parent_child relationships in db.", run.FileID)
		}
	}
}

func TestOrchestrationStoresProblem(t *testing.T) {
	state := config.GetContext("test")
	preTestClear(state)
	db_scratch_pad := context.Background()

	path_descriptor := filehandlers.FileDetails{
		Root:   "./test_files/Erdös_Min_Overlap/run-LLaMEA-gemma4:latest-erdos_min_overlap-0",
		Suffix: "problem.json",
	}

	_, err := filehandlers.ImportProblems(path_descriptor, state, db_scratch_pad)
	if err != nil {
		t.Fatal("Unable to import problem first time: " + err.Error())
	}
	path_descriptor.Root = "./test_files/Erdös_Min_Overlap/run-LLaMEA-llama3.2:latest-erdos_min_overlap-0"
	problem_descriptor, err := filehandlers.ImportProblems(path_descriptor, state, db_scratch_pad)
	if err != nil {
		t.Fatal("Unable to import problem second time: " + err.Error())
	}

	// Check tag population:
	fetched_tags, err := state.DB.GetTagsforProblem(db_scratch_pad, problem_descriptor.ProblemIdentifier.DatabaseID)
	for tag, _ := range problem_descriptor.TagIdentifiers {
		found := slices.Contains(fetched_tags, tag)
		if !found {
			t.Fatalf("Unable to find tag %v in db...", tag)
		}
	}
}

func TestOrchestrationStoresMethod(t *testing.T) {
	state := config.GetContext("test")
	preTestClear(state)
	db_scratch_pad := context.Background()

	path_descriptor := filehandlers.FileDetails{
		Root:   "./test_files/Erdös_Min_Overlap/run-LLaMEA-gemma4:latest-erdos_min_overlap-0",
		Suffix: "method.json",
	}
	method_mapping1, err := filehandlers.ImportMethod(path_descriptor, state, db_scratch_pad)
	if err != nil {
		t.Fatalf("Unable to insert method first time: %v", err)
	}
	method_mapping2, err := filehandlers.ImportMethod(path_descriptor, state, db_scratch_pad)
	if err != nil {
		t.Fatalf("Unable to insert method second time: %v", err)
	}
	if method_mapping1.DatabaseID != method_mapping2.DatabaseID {
		t.Fatal("New id was created for same method.")
	}

}
