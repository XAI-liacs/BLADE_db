-- name: ConnectUserwithExperiment :exec
INSERT INTO user_experiment (user_id, experiment_id)
VALUES (
    $1,
    $2
);

-- name: DeleteExperiment :exec
DELETE FROM user_experiment
WHERE
(
    user_id = $1
    AND
   experiment_id = $2
);
