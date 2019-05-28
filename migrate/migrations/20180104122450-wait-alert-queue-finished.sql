
-- +migrate Up notransaction

-- +migrate StatementBegin
DO
$do$
DECLARE
    max_tries INT := 60;
    c INT := 0;
    n INT := 0;
BEGIN

    DELETE FROM process_alerts WHERE deadline isnull or deadline <= statement_timestamp();

    SELECT COUNT(*)
    FROM process_alerts
    INTO n;

    LOOP
        EXIT WHEN n = 0 OR c = max_tries;

        DELETE FROM process_alerts WHERE deadline isnull or deadline <= statement_timestamp();

        SELECT COUNT(*), c+1
        FROM process_alerts
        INTO n, c;

        PERFORM pg_sleep(1);
    END LOOP;

	IF n != 0 THEN
	    RAISE EXCEPTION 'found active alert jobs';
	END IF;
END
$do$
-- +migrate StatementEnd

-- +migrate Down
