package tests

import (
	filehandlers "XAI-liacs/BLADE_db/file_handlers"
	"XAI-liacs/BLADE_db/internal/config"
	"context"
	"path/filepath"
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

	_, err := filehandlers.ImportProblem(path_descriptor, state, db_scratch_pad)
	if err != nil {
		t.Fatal("Unable to import problem first time: " + err.Error())
	}
	path_descriptor.Root = "./test_files/Erdös_Min_Overlap/run-LLaMEA-llama3.2:latest-erdos_min_overlap-0"
	problem_descriptor, err := filehandlers.ImportProblem(path_descriptor, state, db_scratch_pad)
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

func TestOrchestrationStoresLLM(t *testing.T) {
	state := config.GetContext("test")
	preTestClear(state)
	db_scratch_pad := context.Background()

	path_descriptor := filehandlers.FileDetails{
		Root:   "./test_files/Erdös_Min_Overlap/run-LLaMEA-gemma4:latest-erdos_min_overlap-0",
		Suffix: "llm.json",
	}
	llm_mappings, err := filehandlers.ImportLLM(path_descriptor, state, db_scratch_pad)
	if err != nil {
		t.Fatalf("Unable to insert method first time: %v", err)
	}
	llm_mappings2, err := filehandlers.ImportLLM(path_descriptor, state, db_scratch_pad)
	if err != nil {
		t.Fatalf("Unable to insert method second time: %v", err)
	}
	for index, _ := range llm_mappings {
		if llm_mappings[index].DatabaseID != llm_mappings2[index].DatabaseID {
			t.Fatal("New id was created for same method.")
		}
	}
}

func TestOrchestrationConnectsLLMsMethod(t *testing.T) {
	state := config.GetContext("test")
	preTestClear(state)
	db_scratch_pad := context.Background()

	path_descriptor := filehandlers.FileDetails{
		Root:   "./test_files/Erdös_Min_Overlap/run-LLaMEA-gemma4:latest-erdos_min_overlap-0",
		Suffix: "method.json",
	}
	method_mapping, err := filehandlers.ImportMethod(path_descriptor, state, db_scratch_pad)

	path_descriptor = filehandlers.FileDetails{
		Root:   "./test_files/Erdös_Min_Overlap/run-LLaMEA-gemma4:latest-erdos_min_overlap-0",
		Suffix: "llm.json",
	}
	llm_mappings, err := filehandlers.ImportLLM(path_descriptor, state, db_scratch_pad)
	if err != nil {
		t.Fatalf("Unable to insert method: %v", err)
	}

	err = filehandlers.ConnectMethodLLMs(method_mapping, llm_mappings, state, db_scratch_pad)
	if err != nil {
		t.Fatal("Unable to connect Method with LLM: " + err.Error())
	}
	// Multi-llm to one method.
	path_descriptor = filehandlers.FileDetails{
		Root:   "./test_files/Erdös_Min_Overlap/run-LLaMEA-llama3.2:latest-erdos_min_overlap-0",
		Suffix: "llm.json",
	}
	llm_mappings, err = filehandlers.ImportLLM(path_descriptor, state, db_scratch_pad)
	if err != nil {
		t.Fatalf("Unable to insert method: %v", err)
	}

	err = filehandlers.ConnectMethodLLMs(method_mapping, llm_mappings, state, db_scratch_pad)
	if err != nil {
		t.Fatal("Unable to connect Method with second LLM: " + err.Error())
	}
}

func TestOrchestrationImportRunDescriptor(t *testing.T) {
	state := config.GetContext("test")
	preTestClear(state)
	db_scratch_pad := context.Background()

	path_descriptor := filehandlers.FileDetails{
		Root:   "./test_files/Erdös_Min_Overlap/",
		Suffix: "progress.json",
	}

	progress_data, err := filehandlers.ImportProgress(path_descriptor.Root, state, db_scratch_pad)
	if err != nil {
		t.Fatal("Unable to import progress: " + err.Error())
	}

	for path, run_data := range progress_data.RunData {
		path_descriptor.Root = path
		path_descriptor.Suffix = "method.json"
		method_data, err := filehandlers.ImportMethod(path_descriptor, state, db_scratch_pad)
		if err != nil {
			t.Fatalf("Error importing method in %s, err: %s", filepath.Join(path_descriptor.Root, path_descriptor.Suffix), err.Error())
		}
		path_descriptor.Suffix = "problem.json"
		problem_data, err := filehandlers.ImportProblem(path_descriptor, state, db_scratch_pad)
		if err != nil {
			t.Fatalf("Error importing problem in %s, err: %s", filepath.Join(path_descriptor.Root, path_descriptor.Suffix), err.Error())
		}
		err = filehandlers.ConnectRunDescriptor(run_data.DatabaseID, method_data.DatabaseID, problem_data.ProblemIdentifier.DatabaseID, state, db_scratch_pad)
		if err != nil {
			t.Fatalf("Error connecting descriptor err: %s", err.Error())
		}
	}
}

func TestOrchestrationImportConversationLog(t *testing.T) {
	state := config.GetContext("test")
	preTestClear(state)
	db_scratch_pad := context.Background()

	path_descriptor := filehandlers.FileDetails{
		Root:   "./test_files/Erdös_Min_Overlap",
		Suffix: "conversationlog.json",
	}
	progress_data, err := filehandlers.ImportProgress(path_descriptor.Root, state, db_scratch_pad)
	if err != nil {
		t.Fatal(err)
	}
	for run_path, run := range progress_data.RunData {
		path_descriptor.Root = run_path
		path_descriptor.Suffix = "method.json"
		method_id, err := filehandlers.ImportMethod(path_descriptor, state, db_scratch_pad)
		if err != nil {
			t.Fatal(err)
		}
		path_descriptor.Suffix = "llm.json"
		llms, err := filehandlers.ImportLLM(path_descriptor, state, db_scratch_pad)
		if err != nil {
			t.Fatal(err)
		}
		path_descriptor.Suffix = "conversationlog.jsonl"
		err = filehandlers.ImportConversationLog(path_descriptor, run.DatabaseID, method_id.DatabaseID, llms, state, db_scratch_pad)
		if err != nil {
			t.Fatal(err)
		}
		message_log, err := state.DB.GetConversationLog(db_scratch_pad, run.DatabaseID)
		if len(message_log) != 20 {
			t.Fatal("Not all messanges were imported.")
		}
	}
}

func TestCompleteInjestion(t *testing.T) {
	path, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err.Error())
	}
	path = filepath.Join(path, "/test_files/Erdös_Min_Overlap")

	err = filehandlers.ImportExperiment(path, "test")
	if err != nil {
		t.Fatal(err.Error())
	}
	// Check if commit succeeded.
	state := config.GetContext("test")
	exps, err := state.DB.GetExperiments(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(exps) == 0 {
		t.Fatal("Experiment not imported.")
	}
	runs, err := state.DB.GetRuns(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 {
		t.Fatal("Not all runs imported.")
	}
	for _, run := range runs {
		converation_log, err := state.DB.GetConversationLog(context.Background(), run.ID)
		if err != nil {
			t.Fatal("Cannot get conversation log for run: " + run.ID.String() + " error : " + err.Error())
		}
		if len(converation_log) != 20 {
			t.Fatal("Not all conversation log imported.")
		}
		solutions, err := state.DB.GetSolutionsForRun(context.Background(), run.ID)
		if err != nil {
			t.Fatal("Cannot get solutions for run: " + run.ID.String() + " error : " + err.Error())
		}
		if len(solutions) != 10 {
			t.Fatal("Not all solutions imported.")
		}
		problem_methods, err := state.DB.GetMethodAndProblemForRun(context.Background(), run.ID)
		if err != nil {
			t.Fatal("Cannot get problem_method for run: " + run.ID.String() + " error : " + err.Error())
		}
		if len(problem_methods) != 1 {
			t.Fatal("Not all methods / problems imported.")
		}
		for _, pm := range problem_methods {
			llm, err := state.DB.GetLLMsForMethod(context.Background(), pm.MethodID)
			if err != nil {
				t.Fatal(err.Error())
			}
			if len(llm) == 0 {
				t.Fatal("Cannot get connection between method and llm.")
			}
		}
	}
}

func TestCompleteMultiInjestion(t *testing.T) {
	state := config.GetContext("test")
	preTestClear(state)
	filehandlers.ImportAllExperimentUnder("./test_files", "test")
}
