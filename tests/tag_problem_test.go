package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"anantashahane/BLADE_db/internal/database"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type tagProblemRelation struct {
	tag_id           uuid.UUID
	problem_id       uuid.UUID
	problem_instance map[string]any
}

func prepareTagAndProblem(state config.State) (tag_problem_relation []tagProblemRelation, err error) {

	// Get test files
	tag_problem_relation = make([]tagProblemRelation, 0, 3)
	const base_dir = "test_files/problems"
	files, err := get_file_names(base_dir)
	tag_map := make(map[string]uuid.UUID)
	if err != nil {
		return tag_problem_relation, errors.New("Unable to fetch file names in `" + base_dir + "` , error: " + err.Error())
	}
	// Iterate over test files.
	for _, file := range files {
		if strings.Contains(file, "problem_2") {
			continue
		}
		content, err := read_file(file)
		if err != nil {
			return tag_problem_relation, errors.New("Unable to read file named `" + file + "` , error: " + err.Error())
		}
		data, err := json.Marshal(content["config"])
		if err != nil {
			return tag_problem_relation, errors.New("Unable to serialise config for file `" + file + "` , error: " + err.Error())
		}
		problem_id, err := state.DB.CreateProblem(context.Background(), database.CreateProblemParams{
			ID:           uuid.New(),
			Name:         content["name"].(string),
			Prompt:       content["prompt"].(string),
			Evaluator:    content["evaluator"].(string),
			Minimisation: content["minimisation"].(bool),
			Config:       pqtype.NullRawMessage{RawMessage: data, Valid: err == nil},
		})
		if err != nil {
			if !strings.Contains(err.Error(), "duplicate key value violates unique constraint \"unique_problem\"") {
				return tag_problem_relation, errors.New("Unable to populate tags `" + file + "` , error: " + err.Error())
			}
		}
		tags, ok := content["tags"].([]any)
		if !ok {
			return tag_problem_relation, errors.New("Cannot cascade tag interface into []interface.")
		}

		for _, tag := range tags {
			tag_id, err := state.DB.CreateTag(context.Background(), database.CreateTagParams{ID: uuid.New(), Tag: tag.(string)})
			if err != nil {
				if !strings.Contains(err.Error(), "duplicate key value violates unique constraint \"tag_tag_key\"") {
					return tag_problem_relation, errors.New("Unable to populate tags `" + file + "` , error: " + err.Error())
				}
			} else {
				tag_map[tag.(string)] = tag_id
			}
			if tag_id != uuid.Nil && problem_id != uuid.Nil {
				tag_problem_relation = append(tag_problem_relation, tagProblemRelation{tag_id: tag_map[tag.(string)], problem_id: problem_id, problem_instance: content})
			}
		}
	}
	return tag_problem_relation, nil
}

func TestRelationTableTagProblemInsertion(t *testing.T) {
	state := config.GetContext("test")
	tag_problem_relations, err := prepareTagAndProblem(state)
	problem_map := make(map[uuid.UUID]tagProblemRelation)
	if err != nil {
		t.Fatal(err.Error())
	}
	// Connect tables
	for _, tag_problem := range tag_problem_relations {
		err = state.DB.ConnectTagProblem(context.Background(), database.ConnectTagProblemParams{
			TagID:     tag_problem.tag_id,
			ProblemID: tag_problem.problem_id,
		})
		if err != nil {
			t.Fatalf("Unable to connect problem %s with tag %s, error: %s", tag_problem.tag_id, tag_problem.problem_id, err.Error())
		} else {
			problem_map[tag_problem.problem_id] = tag_problem
		}
	}

	//Test validity (Read).
	for k, v := range problem_map {
		problem_tags, err := state.DB.GetTagsforProblem(context.Background(), k)
		if err != nil {
			t.Fatalf("Unable to fetch tags for problem id: %s, error: %s", k, err.Error())
		}
		for _, fetched_tag := range problem_tags {
			inserted_tags := v.problem_instance["tags"].([]any)
			mapped_properly := false
			for _, inserted_tag := range inserted_tags {
				if inserted_tag.(string) == fetched_tag {
					mapped_properly = true
				}
			}
			if !mapped_properly {
				t.Fatalf("Tag %s fetch mis-matched.", fetched_tag)
			}
		}
	}
}

func TestRelationTableTagProblemRejectsUnknownIDs(t *testing.T) {
	state := config.GetContext("test")
	err := state.DB.CreateRunSolution(context.Background(), database.CreateRunSolutionParams{
		RunID:      uuid.New(),
		SolutionID: uuid.New(),
	})
	if err == nil {
		t.Fatalf("DB accepted unknown keys.")
	}
}
