import { DateTime, Duration, Interval } from 'luxon'
import { Chance } from 'chance'
const c = new Chance()

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

// randInterval creates a random interval in the future.
export function randInterval(): Interval {
  const now = DateTime.utc()
  const start = now
    .plus({ hour: 1, days: c.floating({ min: 0, max: 3 }) })
    .startOf('minute') // always start in the future
  const end = start
    .plus({ days: c.floating({ min: 2, max: 4 }) })
    .startOf('minute') // always at least 1 hour long

  return Interval.fromDateTimes(start, end)
}

// randDTWithinInterval will return a random DateTime within the provided interval.
export function randDTWithinInterval(ivl: Interval): DateTime {
  const startMSec = ivl.start.toMillis()
  const endMSec = ivl.end.plus({ minutes: -1 }).toMillis()
  return DateTime.fromMillis(
    c.integer({ min: startMSec, max: endMSec }),
  ).startOf('minute')
}

// randDT returns a random DateTime.
export function randDT({
  min,
  max,
}: {
  min?: DateTime
  max?: DateTime
}): DateTime {
  if (!min) min = DateTime.utc().plus({ minutes: 15 })
  if (!max) max = min.plus({ days: 7 })

  return randDTWithinInterval(Interval.fromDateTimes(min, max))
}

// randSubInterval creates a random Interval within an Interval.
export function randSubInterval(
  ivl: Interval,
  { min, max }: { min?: Duration; max?: Duration } = {},
): Interval {
  if (!min) min = Duration.fromObject({ minutes: 15 })
  if (!max) max = Duration.fromObject({ hour: 1 })
  if (ivl.toDuration() <= min) {
    throw new Error('interval too small for min value')
  }

  let a = randDTWithinInterval(ivl)
  let b = randDTWithinInterval(ivl)
  let diff = a.diff(b)
  if (diff.valueOf() < 0) diff = diff.negate()

  while (diff < min || diff > max) {
    a = randDTWithinInterval(ivl)
    b = randDTWithinInterval(ivl)
    diff = a.diff(b)
    if (diff.valueOf() < 0) diff = diff.negate()
  }

  if (a < b) {
    return Interval.fromDateTimes(a, b)
  }

  return Interval.fromDateTimes(b, a)
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
    before(() => {
      cy.clearCookie('goalert_session.2')
      resetQuery().then((query) =>
        cy.task('engine:stop').sql(query).task('engine:start'),
      )
    })
    it('reset db', () => {}) // required due to mocha skip bug

    if (!skipLogin) {
      before(() => cy.resetConfig()[adminLogin ? 'adminLogin' : 'login']())
      it(adminLogin ? 'admin login' : 'login', () => {}) // required due to mocha skip bug
    }

    describe(screenName(), () => fn(screen()))
  })
}
