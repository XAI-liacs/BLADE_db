package filehandlers

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

/*
 * ID mapping struct is used to map the real id with that of db_id.
 * The database in tables message, problem, method, tag & llm will return ID for the already available rows,
 * that are exact copy of the one being inserted.
 */
type ID_Mapping struct {
	FileID     uuid.UUID
	DatabaseID uuid.UUID
}

type ProgressImportData struct {
	Experiment ID_Mapping
	RunData    map[string]ID_Mapping //Directory -> DB_ID mapping.
}

/*
 * Imports experiment under a directory. Updates experiment, run, and experiment_run tables, while returning the associated db ids.
 *
 * ## Args
 *
 * - `project_path : string`: Path to the root of the experiment, i.e., path where `progress.json` exist for a given experiment.
 * - `state: State struct`: Contains db related functions.
 * - `db_scratch_pad: context.Context`: Allows for cancellable update of the database, if an error is encountered, context.cancel().
 *
 * ## Returns
 * - `progress_ids: ProgressImportData` for all future references.
 */
func ImportProgress(project_path string, state config.State, db_scratch_pad context.Context) (progress_ids ProgressImportData, err error) {

	// Step 1: Check if provided path is actually pwd for experiment (root dir must contain `progress.json`)
	found_progress_json := false
	files_in_directory, err := os.ReadDir(project_path)
	if err != nil {
		return progress_ids, err
	}
	for _, file := range files_in_directory {
		if file.Name() == "progress.json" {
			found_progress_json = true
			break
		}
	}

	if !found_progress_json {
		file_name := filepath.Join(project_path, "progress")
		return progress_ids, fmt.Errorf("File %v not found.", file_name)
	}

	path_details := FileDetails{
		Root:   project_path,
		Suffix: "progress.json",
	}

	// Injest progress.
	progress := Progress{}
	progress_data, err := progress.Read(path_details)
	if err != nil {
		return progress_ids, err
	}
	// Add experiment into DB
	id, err := state.DB.CreateExperiment(db_scratch_pad, progress_data.progressData)
	if err != nil {
		return progress_ids, err
	}

	progress_ids.Experiment = ID_Mapping{
		FileID:     progress_data.progressData.ID,
		DatabaseID: id,
	}

	progress_ids.RunData = make(map[string]ID_Mapping)

	// Injest runs.
	for sub_directory, run := range progress_data.runData {
		run_id, err := state.DB.CreateRun(db_scratch_pad, run)
		if err != nil {
			return progress_ids, err
		}
		absolute_path := filepath.Join(path_details.Root, sub_directory)
		if err != nil {
			return progress_ids, err
		}
		progress_ids.RunData[absolute_path] = ID_Mapping{
			FileID:     run.ID,
			DatabaseID: run_id,
		}
	}

	// Join run with experiment:
	for _, run_id := range progress_ids.RunData {
		err = state.DB.CreateExperimentRun(db_scratch_pad, database.CreateExperimentRunParams{
			ExperimentID: progress_ids.Experiment.DatabaseID,
			RunID:        run_id.DatabaseID,
		})
		if err != nil {
			return
		}
	}
	return
}

/*
 * Imports Solution under a directory. Updates solution, solution_parent_child and run_solution tables, while returning the associated db ids.
 *
 * ## Args
 *
 * - `path_descriptor : FileDetails`: Path to the root of the run, with Suffix=log.jsonl.
 * - `run_id : uuid.UUID[16]`: DB ID of the imported run; associated with the solution (see key, value pair of ProgressImportData.RunData)
 * - `state: State struct`: Contains db related functions.
 * - `db_scratch_pad: context.Context`: Allows for cancellable update of the database, if an error is encountered, context.cancel().
 *
 * ## Returns
 * - `solution_ids: []ID_Mapping` for all future references.
 */
func ImportSolution(path_descriptor FileDetails, run_id uuid.UUID, state config.State, db_scratch_pad context.Context) (solution_ids []ID_Mapping, err error) {
	// Import all the solutions.
	sl := SolutionLog{}

	solutions, err := sl.Read(path_descriptor)
	if err != nil {
		return
	}
	solution_ids = make([]ID_Mapping, 0)

	solution_id_mappings := make(map[uuid.UUID]uuid.UUID) // Solution file to db id mapping to build childparent table.
	for _, solution := range solutions {
		fitness_raw_data, err := json.Marshal(solution.Fitness)
		if err != nil {
			return solution_ids, fmt.Errorf("Unable to serialise fitness data (%v), error: %s", solution.Fitness, err.Error())
		}
		metadata, err := json.Marshal(solution.Metadata)
		if err != nil {
			return solution_ids, fmt.Errorf("Unable to serialise metadata data (%v), error: %s", solution.Metadata, err.Error())
		}
		id, err := state.DB.CreateSolution(db_scratch_pad, database.CreateSolutionParams{
			ID:          solution.ID,
			Name:        sql.NullString{String: solution.Name, Valid: true},
			Description: sql.NullString{String: solution.Description, Valid: true},
			Generation:  sql.NullInt32{Int32: int32(OptionalUnwrap(solution.Generation)), Valid: solution.Generation != nil},
			Code:        sql.NullString{String: solution.Code, Valid: true},
			Metadata:    pqtype.NullRawMessage{RawMessage: metadata, Valid: len(metadata) != 0},
			Fitness:     pqtype.NullRawMessage{RawMessage: fitness_raw_data, Valid: len(fitness_raw_data) != 0},
		})
		if err != nil {
			return solution_ids, err
		}

		err = state.DB.CreateRunSolution(db_scratch_pad, database.CreateRunSolutionParams{
			RunID:      run_id,
			SolutionID: id,
		})
		if err != nil {
			return solution_ids, err
		}

		solution_ids = append(solution_ids, ID_Mapping{
			FileID:     solution.ID,
			DatabaseID: id,
		})
		solution_id_mappings[solution.ID] = id
	}

	// Populate parent-child relationship.
	for _, solution := range solutions {
		for _, parent_id := range solution.ParentIDs {
			db_pid := solution_id_mappings[parent_id]
			db_child_id := solution_id_mappings[solution.ID]
			state.DB.CreateParentChild(db_scratch_pad, database.CreateParentChildParams{
				ParentID: db_pid,
				ChildID:  db_child_id,
			})
			if err != nil {
				return solution_ids, err
			}
		}
	}
	return solution_ids, nil
}

/*
 * Imports Problem under a directory. Updates problem, tags and problem_tags tables, while returning the associated db ids.
 *
 * ## Args
 *
 * - `path_descriptor : FileDetails`: Path to the root of the run, with Suffix=problem.json.
 * - `state: State struct`: Contains db related functions.
 * - `db_scratch_pad: context.Context`: Allows for cancellable update of the database, if an error is encountered, context.cancel().
 *
 * ## Returns
 * - `solution_ids: []ID_Mapping` for all future references.
 */
type ProblemTag struct {
	ProblemIdentifier ID_Mapping
	TagIdentifiers    map[string]ID_Mapping //Tag to problem_id
}

func ImportProblems(path_descriptor FileDetails, state config.State, db_scratch_pad context.Context) (problem_descriptor ProblemTag, err error) {
	p := Problem{}
	problem_content, err := p.Read(path_descriptor)
	if err != nil {
		return
	}
	problem_id, err := state.DB.CreateProblem(db_scratch_pad, problem_content.Problem)
	if err != nil {
		return
	}

	problem_descriptor.ProblemIdentifier = ID_Mapping{
		FileID:     problem_content.Problem.ID,
		DatabaseID: problem_id,
	}

	problem_descriptor.TagIdentifiers = make(map[string]ID_Mapping)
	for _, tag := range problem_content.Tags {
		id, err := state.DB.CreateTag(db_scratch_pad, tag)
		if err != nil {
			return problem_descriptor, err
		}
		err = state.DB.ConnectTagProblem(db_scratch_pad, database.ConnectTagProblemParams{
			ProblemID: problem_id,
			TagID:     id,
		})
		if err != nil {
			return problem_descriptor, err
		}
		problem_descriptor.TagIdentifiers[tag.Tag] = ID_Mapping{
			FileID:     tag.ID,
			DatabaseID: id,
		}
	}
	return problem_descriptor, nil
}

/*
 * Imports method under a directory. Updates method table, while returning the associated db ids.
 *
 * ## Args
 *
 * - `path_descriptor : FileDetails`: Path to the root of the run, with Suffix=method.json.
 * - `state: State struct`: Contains db related functions.
 * - `db_scratch_pad: context.Context`: Allows for cancellable update of the database, if an error is encountered, context.cancel().
 *
 * ## Returns
 * - `solution_ids: []ID_Mapping` for all future references.
 */

func ImportMethod(path_descriptor FileDetails, state config.State, db_scratch_pad context.Context) (method_id_mapping ID_Mapping, err error) {
	m := Method{}
	method_content, err := m.Read(path_descriptor)
	if err != nil {
		return
	}
	method_id, err := state.DB.CreateMethod(db_scratch_pad, method_content)
	if err != nil {
		return
	}

	method_id_mapping = ID_Mapping{
		FileID:     method_content.ID,
		DatabaseID: method_id,
	}
	return method_id_mapping, nil
}
