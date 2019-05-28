
-- +migrate Up

DROP TYPE IF EXISTS enum_user_role;
CREATE TYPE enum_user_role as ENUM (
    'unknown',
    'user',
    'admin'
);

ALTER TABLE goalert_user 
    RENAME TO users;

UPDATE users
    SET bio = CASE
        WHEN bio IS NULL THEN ''
        ELSE bio
    END;
ALTER TABLE users ALTER COLUMN bio SET NOT NULL;
ALTER TABLE users ALTER COLUMN bio SET DEFAULT '';

UPDATE users
    SET role = CASE
        WHEN role IS NULL THEN 'user'
        WHEN role='' THEN 'user'
        ELSE role
    END;
ALTER TABLE users ALTER COLUMN role TYPE enum_user_role USING role::enum_user_role;
ALTER TABLE users ALTER COLUMN role SET NOT NULL;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'unknown'::enum_user_role;

ALTER TABLE users ALTER COLUMN email SET NOT NULL;

ALTER TABLE users DROP COLUMN login;
ALTER TABLE users DROP COLUMN schedule_color;
ALTER TABLE users DROP COLUMN time_zone;
ALTER TABLE users DROP COLUMN title;

ALTER TABLE users ADD COLUMN name TEXT;
UPDATE users SET name = LTRIM(CASE
        WHEN first_name IS NULL THEN ''
        ELSE first_name
    END||CASE
        WHEN last_name IS NULL THEN ''
        WHEN last_name='' THEN ''
        ELSE ' '||last_name
    END);
ALTER TABLE users ALTER COLUMN name SET NOT NULL;

ALTER TABLE users DROP COLUMN last_name;
ALTER TABLE users DROP COLUMN first_name;


-- +migrate Down

ALTER TABLE users RENAME TO goalert_user;

ALTER TABLE goalert_user ALTER COLUMN bio DROP NOT NULL;
ALTER TABLE goalert_user ALTER COLUMN role DROP NOT NULL;
ALTER TABLE goalert_user ALTER COLUMN role TYPE text USING role::text;
ALTER TABLE goalert_user ALTER COLUMN role SET DEFAULT 'user'::text;

ALTER TABLE goalert_user ADD COLUMN login TEXT UNIQUE;
ALTER TABLE goalert_user ADD COLUMN schedule_color TEXT;
ALTER TABLE goalert_user ADD COLUMN time_zone TEXT;
ALTER TABLE goalert_user ADD COLUMN title TEXT;
ALTER TABLE goalert_user ADD COLUMN last_name TEXT DEFAULT '';
ALTER TABLE goalert_user ADD COLUMN first_name TEXT DEFAULT '';

UPDATE goalert_user SET first_name = name;
ALTER TABLE goalert_user DROP COLUMN name;

DROP TYPE enum_user_role;
