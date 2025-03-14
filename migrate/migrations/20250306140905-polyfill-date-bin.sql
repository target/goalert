-- +migrate Up
DO $$
BEGIN
    IF current_setting('server_version_num')::int < 140000 THEN
        CREATE OR REPLACE FUNCTION public.date_bin(p_bin_interval interval, p_ts timestamp with time zone, p_origin timestamp with time zone )
            RETURNS timestamp with time zone
            LANGUAGE SQL
            IMMUTABLE AS $func$
            SELECT
                to_timestamp (
floor((extract(epoch FROM p_ts ) - extract(epoch FROM p_origin ) ) / extract(epoch FROM p_bin_interval ) ) * extract(epoch FROM p_bin_interval ) + extract(epoch FROM p_origin )
                );
    $func$;
END IF;
END
$$;

-- +migrate Down
DROP FUNCTION IF EXISTS public.date_bin(interval, timestamp with time zone, timestamp with time zone);

