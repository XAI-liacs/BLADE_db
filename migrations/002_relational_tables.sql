-- +goose Up
CREATE TABLE ExperimentRun(
    experiment_id UUID NOT NULL 
        REFERENCES Experiment(id),
    run_id UUID NOT NULL 
        REFERENCES Run(id),
    UNIQUE (experiment_id, run_id)
);

CREATE TABLE ConversationLog(
    run_id UUID NOT NULL 
        REFERENCES Run(id),
    message_id UUID NOT NULL 
        REFERENCES Message(id),
    UNIQUE (run_id, message_id)
);

CREATE TABLE Sender (
    message_id UUID NOT NULL
        REFERENCES Message(id),
    llm_id UUID
        REFERENCES LLM(id),
    method_id UUID
        REFERENCES Method(id),
    UNIQUE (message_id, llm_id, method_id),
    CHECK (
        (llm_id IS NOT NULL AND method_id IS NULL)
        OR
        (llm_id IS NULL AND method_id IS NOT NULL)
    )
);

CREATE TABLE ProblemTags(
    tag_id UUID NOT NULL
        REFERENCES Tag(id),
    problem_id UUID NOT NULL
        REFERENCES Problem(id),
    UNIQUE (tag_id, problem_id)
);

CREATE TABLE MethodLLM(
    llm_id UUID NOT NULL
        REFERENCES LLM(id),
    method_id UUID NOT NULL
        REFERENCES Method(id),
    UNIQUE (llm_id, method_id)
);

CREATE TABLE RunDescriptor(
    run_id UUID NOT NULL
        REFERENCES Run(id),
    method_id UUID NOT NULL
        REFERENCES Method(id),
    problem_id UUID NOT NULL
        REFERENCES Problem(id),
    UNIQUE (run_id, method_id, problem_id)
);

CREATE TABLE RunSolution(
    run_id UUID NOT NULL
        REFERENCES Run(id),
    solution_id UUID NOT NULL
        REFERENCES Solution(id),
    UNIQUE (run_id, solution_id)
);

CREATE TABLE ParentChild(
    parent_id UUID NOT NULL
        REFERENCES Solution(id),
    child_id UUID NOT NULL
        REFERENCES Solution(id),
    UNIQUE (parent_id, child_id)
);

-- +goose Down
DROP Table ExperimentRun;
DROP Table ConversationLog;
DROP Table Sender;
DROP Table ProblemTags;
DROP Table MethodLLM;
DROP Table RunDescriptor;
DROP Table RunSolution;
DROP Table ParentChild;