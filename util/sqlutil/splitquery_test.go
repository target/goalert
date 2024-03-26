package sqlutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const adv = `
CREATE OR REPLACE FUNCTION public.fn_enforce_alert_limit()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'unacked_alerts_per_service';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM alerts
    WHERE service_id = NEW.service_id AND "status" = 'triggered';

    IF val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='unacked_alerts_per_service_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$function$
;
`

func TestSplitQuery(t *testing.T) {
	assert.Equal(t, []string{"foobar"}, SplitQuery("foobar"))
	assert.Equal(t, []string{"foo", "bar"}, SplitQuery("foo;bar"))
	assert.Equal(t, []string{"foo", "bar"}, SplitQuery("foo;bar;"))
	assert.Equal(t, []string{"foo$$; $$bar", "baz"}, SplitQuery("foo$$; $$bar;baz"))

	assert.Len(t, SplitQuery(adv), 1)
}
