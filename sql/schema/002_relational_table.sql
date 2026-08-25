-- +goose Up

CREATE TABLE solution_parent_child (
    parent_id UUID NOT NULL REFERENCES solution(id),
    child_id  UUID NOT NULL REFERENCES solution(id),
    CONSTRAINT parent_child UNIQUE (parent_id, child_id)
);

CREATE TABLE experiment_run(
    experiment_id UUID NOT NULL REFERENCES experiment(id),
    run_id UUID NOT NULL REFERENCES run(id),
    CONSTRAINT experiment_run_relation UNIQUE (experiment_id, run_id)
);

CREATE TABLE run_solution(
    run_id UUID NOT NULL REFERENCES run(id),
    solution_id UUID NOT NULL REFERENCES solution(id),
    CONSTRAINT run_solutions_relation UNIQUE (run_id, solution_id)
);

CREATE TABLE tag_problem(
    tag_id UUID NOT NULL REFERENCES tag(id),
    problem_id UUID NOT NULL REFERENCES problem(id),
    CONSTRAINT tag_problem_relation UNIQUE (tag_id, problem_id)
);

CREATE TABLE method_llm(
    method_id UUID NOT NULL REFERENCES method(id),
    llm_id UUID NOT NULL REFERENCES llm(id),
    CONSTRAINT method_llm_relation UNIQUE (method_id, llm_id)
);

CREATE TABLE run_descriptor(
    run_id UUID NOT NULL REFERENCES run(id),
    method_id UUID NOT NULL REFERENCES method(id),
    problem_id UUID NOT NULL REFERENCES problem(id),
    CONSTRAINT run_descriptor_relation UNIQUE (run_id, method_id, problem_id)
);

CREATE TABLE conversation_log(
    run_id UUID NOT NULL REFERENCES run(id),
    message_id UUID NOT NULL REFERENCES message(id),
    method_id UUID NOT NULL REFERENCES method(id),
    llm_id UUID NOT NULL REFERENCES llm(id),
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT conversation_log_relation UNIQUE (run_id, message_id, method_id, llm_id, created_at)
);

-- +goose Down

DROP TABLE solution_parent_child;
DROP TABLE experiment_run;
DROP TABLE run_solution;
DROP TABLE tag_problem;
DROP TABLE method_llm;
DROP TABLE run_descriptor;
DROP TABLE conversation_log;
