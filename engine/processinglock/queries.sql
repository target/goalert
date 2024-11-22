-- name: ProcSharedAdvisoryLock :one
SELECT
    pg_try_advisory_xact_lock_shared($1) AS lock_acquired;

-- name: ProcAcquireModuleLock :one
SELECT
    version
FROM
    engine_processing_versions
WHERE
    type_id = $1
FOR UPDATE
    NOWAIT;

-- name: ProcReadModuleVersion :one
SELECT
    version
FROM
    engine_processing_versions
WHERE
    type_id = $1;

-- name: ProcSaveState :exec
UPDATE
    engine_processing_versions
SET
    state = $2
WHERE
    type_id = $1;

-- name: ProcLoadState :one
SELECT
    state
FROM
    engine_processing_versions
WHERE
    type_id = $1;

