const bcrypt = require('bcryptjs')

const profile = require('../cypress/fixtures/profile.json')
const profileAdmin = require('../cypress/fixtures/profileAdmin.json')
const users = require('../cypress/fixtures/users.json').concat(
  profile,
  profileAdmin,
)

const ids = users.map(u => `'${u.id}'`).join(',')
const userVals = users
  .map(u => `('${u.id}','${u.name}','${u.email}','${u.role}')`)
  .join(',\n')
const pwHash = bcrypt.hashSync(profile.password, bcrypt.genSaltSync(1))
const pwHashAdmin = bcrypt.hashSync(
  profileAdmin.password,
  bcrypt.genSaltSync(1),
)

console.log(
  `
delete from users where id in (${ids});

insert into users (id, name, email, role)
values
  ${userVals}
;
insert into auth_basic_users (user_id, username, password_hash)
values
  ('${profile.id}', '${profile.username}', '${pwHash}'),
  ('${profileAdmin.id}', '${profileAdmin.username}', '${pwHashAdmin}');
`,
)
