-- +migrate Up
-- we're fixing a bug where the gql_api_keys_name_key unique index does not account for soft deletes
ALTER TABLE gql_api_keys
    DROP CONSTRAINT gql_api_keys_name_key;

CREATE UNIQUE INDEX gql_api_keys_name_key ON public.gql_api_keys(name)
WHERE
    deleted_at IS NULL;

-- +migrate Down
DROP INDEX gql_api_keys_name_key;

ALTER TABLE ONLY public.gql_api_keys
    ADD CONSTRAINT gql_api_keys_name_key UNIQUE (name);

