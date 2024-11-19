-- name: ProcSharedAdvisoryLock :one
SELECT
    pg_try_advisory_xact_lock_shared($1) AS lock_acquired;

-- name: ProcAcquireModuleLockNoWait :one
SELECT
    1
FROM
    engine_processing_versions
WHERE
    type_id = $1
    AND version = $2
FOR UPDATE
    NOWAIT;

-- name: ProcAcquireModuleSharedLock :one
SELECT
    1
FROM
    engine_processing_versions
WHERE
    type_id = $1
    AND version = $2 FOR SHARE;

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

