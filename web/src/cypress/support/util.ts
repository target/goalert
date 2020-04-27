declare global {
  type ScreenFormat = 'mobile' | 'tablet' | 'widescreen'
}

let _resetQuery = ''
function resetQuery(): Cypress.Chainable<string> {
  if (_resetQuery) return cy.wrap(_resetQuery)

  let users: Profile[] = []
  let profile: Profile
  let profileAdmin: Profile
  cy.fixture('users').then((u) => {
    users = users.concat(u)
  })
  cy.fixture('profile').then((p) => {
    profile = p
    users = users.concat(p)
  })
  cy.fixture('profileAdmin').then((p) => {
    profileAdmin = p
    users = users.concat(p)
  })
  return cy.then(() => {
    const ids = users.map((u) => `'${u.id}'`).join(',')
    const userVals = users
      .map((u) => `('${u.id}','${u.name}','${u.email}','${u.role}')`)
      .join(',\n')

    _resetQuery = `
truncate table escalation_policies, rotations, schedules, users CASCADE;
select setval(pg_get_serial_sequence('alerts', 'id'), coalesce(max(id), 0)+10000 , false) from alerts;
delete from users where id in (${ids});

insert into users (id, name, email, role)
values
  ${userVals}
;
insert into auth_basic_users (user_id, username, password_hash)
values
  ('${profile.id}', '${profile.username}', '${profile.passwordHash}'),
  ('${profileAdmin.id}', '${profileAdmin.username}', '${profileAdmin.passwordHash}');
`
    return _resetQuery
  })
}

export function screen(): ScreenFormat {
  const width = Cypress.config().viewportWidth
  if (width < 600) return 'mobile'
  if (width < 960) return 'tablet'

  return 'widescreen'
}

export function screenName(): string {
  switch (screen()) {
    case 'mobile':
      return 'Mobile'
    case 'tablet':
      return 'Tablet'
  }

  return 'Wide'
}

export function testScreen(
  label: string,
  fn: (screen: ScreenFormat) => void,
  skipLogin = false,
  adminLogin = false,
): void {
  describe(label, () => {
    before(() =>
      resetQuery().then((query) =>
        cy.task('engine:stop').sql(query).task('engine:start'),
      ),
    )
    it('reset db', () => {}) // required due to mocha skip bug

    if (!skipLogin) {
      before(() => cy.resetConfig()[adminLogin ? 'adminLogin' : 'login']())
      it(adminLogin ? 'admin login' : 'login', () => {}) // required due to mocha skip bug
    }

    describe(screenName(), () => fn(screen()))
  })
}
