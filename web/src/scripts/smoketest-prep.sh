#!/bin/sh
PWHASH='$2a$06$EWBXfZ.CA1LBPEanxdKoJefU5UGKic39p/69ByppqhPP3tPki2YuO'
USERID='34050a34-466a-4ee6-bc54-d9df243c6463'
USERID2='ADCBDDF6-04C5-489A-A50B-1AC9281A73ED'

if [ -z "$GOALERT" ]
then
  GOALERT=bin/goalert
fi

if [ -z "$BASE_URL" ]
then
  BASE_URL=http://localhost:3040
fi

if [ -z "$DB_URL" ]
then
  DB_URL=postgres://goalert@localhost:5432?sslmode=disable
fi

$GOALERT migrate --db-url=$DB_URL
echo "{\"General\":{\"PublicURL\": \"$BASE_URL\"}}" | $GOALERT set-config  --allow-empty-data-encryption-key --db-url=$DB_URL

psql -d $DB_URL <<EOF

delete from users where id = '$USERID' OR id = '$USERID2';

SELECT setval(pg_get_serial_sequence('alerts', 'id'), coalesce(max(id), 0)+10000 , false) FROM alerts;

$(node scripts/insert-users.js)

insert into users (id, name, email, role)
values
  (
    '$USERID',
    'User McUserFace',
    'test@example.com',
    'user'
  ),
  (
    '$USERID2',
    'Other UserFace',
    'test2@example.com',
    'admin'
  );

insert into auth_basic_users (user_id, username, password_hash)
values
  (
    '$USERID',
    'smoketest',
    '$PWHASH'
  );

insert into auth_basic_users (user_id, username, password_hash)
values
  (
    '$USERID2',
    'smoketestadmin',
    '$PWHASH'
  );
EOF
