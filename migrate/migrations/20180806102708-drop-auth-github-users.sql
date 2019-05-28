
-- +migrate Up

DROP TABLE auth_github_users;
DROP FUNCTION fn_insert_github_user();

-- +migrate Down
CREATE TABLE auth_github_users (
    user_id uuid REFERENCES users(id) ON DELETE CASCADE PRIMARY KEY,
    github_id text NOT NULL UNIQUE
);
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
