
-- +migrate Up

CREATE TABLE auth_user_sessions (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    user_agent TEXT NOT NULL DEFAULT '',
    user_id UUID REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE auth_subjects (
    provider_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (provider_id, subject_id)
)
WITH (fillfactor = 80);

LOCK auth_github_users, auth_basic_users;

INSERT INTO auth_subjects (provider_id, subject_id, user_id)
SELECT 'github', github_id, user_id
FROM auth_github_users;

INSERT INTO auth_subjects (provider_id, subject_id, user_id)
SELECT 'basic', username, user_id
FROM auth_basic_users;


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_insert_basic_user() RETURNS TRIGGER AS
$$
BEGIN

    INSERT INTO auth_subjects (provider_id, subject_id, user_id)
    VALUES ('basic', NEW.username, NEW.user_id)
    ON CONFLICT (provider_id, subject_id) DO UPDATE
    SET user_id = NEW.user_id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_insert_github_user() RETURNS TRIGGER AS
$$
BEGIN

    INSERT INTO auth_subjects (provider_id, subject_id, user_id)
    VALUES ('github', NEW.github_id::text, NEW.user_id)
    ON CONFLICT (provider_id, subject_id) DO UPDATE
    SET user_id = NEW.user_id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_insert_github_user
AFTER INSERT ON auth_github_users
FOR EACH ROW
EXECUTE PROCEDURE fn_insert_github_user();

CREATE TRIGGER trg_insert_basic_user
AFTER INSERT ON auth_basic_users
FOR EACH ROW
EXECUTE PROCEDURE fn_insert_basic_user();

-- +migrate Down

DROP TRIGGER trg_insert_github_user ON auth_github_users;
DROP FUNCTION fn_insert_github_user();

DROP TRIGGER trg_insert_basic_user ON auth_basic_users;
DROP FUNCTION fn_insert_basic_user();

INSERT INTO auth_github_users (github_id, user_id)
SELECT subject_id, user_id
FROM auth_subjects
WHERE provider_id = 'github'
ON CONFLICT (github_id) DO UPDATE
SET user_id = excluded.user_id;

DROP TABLE auth_subjects;
DROP TABLE auth_user_sessions;
