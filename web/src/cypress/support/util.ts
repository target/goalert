import { DateTime, Duration, Interval } from 'luxon'
import { Chance } from 'chance'
import users from '../fixtures/users.json'
import profile from '../fixtures/profile.json'
import profileAdmin from '../fixtures/profileAdmin.json'
import type { ExpFlag } from '../../expflag'

const allUsers = users.concat(profile, profileAdmin)
const c = new Chance()

declare global {
  type ScreenFormat = 'mobile' | 'tablet' | 'widescreen'
}
const ids = allUsers.map((u) => `'${u.id}'`).join(',')
const userVals = allUsers
  .map((u) => `('${u.id}','${u.name}','${u.email}','${u.role}')`)
  .join(',\n')

const resetQuery = `
truncate table escalation_policies, rotations, schedules, alert_metrics, alert_logs, users CASCADE;
select setval(pg_get_serial_sequence('alerts', 'id'), coalesce(max(id), 0)+10000 , false) from alerts;
delete from users where id in (${ids});

update engine_processing_versions set state = '{}' where type_id = 'metrics';

insert into users (id, name, email, role)
values
  ${userVals}
;
insert into auth_basic_users (user_id, username, password_hash)
values
  ('${profile.id}', '${profile.username}', '${profile.passwordHash}'),
  ('${profileAdmin.id}', '${profileAdmin.username}', '${profileAdmin.passwordHash}');
`

// pathPrefix will return the path prefix for the current environment
//
// Trailing `/` is removed, so it is safe to use `pathPrefix + '/foo'`.
export function pathPrefix(): string {
  return new URL(Cypress.config().baseUrl || '').pathname.replace(/\/$/, '')
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
  min?: DateTime | string | null
  max?: DateTime | string | null
}): DateTime {
  if (typeof min === 'string') min = DateTime.fromISO(min)
  if (typeof max === 'string') max = DateTime.fromISO(max)

  if (!min && max) min = max.minus({ days: 7 })
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

let testN = 0

let loginFn: (() => void) | null = null

export function login(): void {
  if (!loginFn) throw new Error('only valid within a testScreen block')

  loginFn()
}

export function testScreen(
  label: string,
  fn: (screen: ScreenFormat) => void,
  skipLogin = false,
  adminLogin = false,
  expFlags: ExpFlag[] = [],
): void {
  after(() => {
    loginFn = null
  })

  if (!skipLogin) {
    const sessID = { n: testN++, ts: new Date() }
    loginFn = () => {
      cy.session(sessID, () => {
        cy.resetConfig()[adminLogin ? 'adminLogin' : 'login']()
      })
    }
  } else {
    loginFn = null
  }

  describe(label, () => {
    before(function () {
      cy.task('check:abort').then((abort) => {
        if (abort) {
          this.skip()
        }
      })
      // Ensure we cancel any in-flight requests by navigating to a blank page before setup.
      //
      // Fixes issues with stale cookies being set from the previous test.
      cy.intercept('/_cy_test_reset', '<html></html>')
      cy.visit('/_cy_test_reset')

      cy.task('engine:stop')
        .sql(resetQuery)
        .task('db:resettime')
        .task('engine:setexpflags', expFlags)
        .task('engine:start')
    })
    it('reset db', () => {}) // required due to mocha skip bug

    if (!skipLogin) {
      if (!loginFn) throw new Error('loginFn not set')
      beforeEach(loginFn)
    }

    describe(screenName(), () => fn(screen()))
  })
}

// testScreenWithFlags is a convenience function for testing a screen with
// different experimental flags.
//
// It is equivalent to calling testScreen(label, fn, false, false, expFlags)
export function testScreenWithFlags(
  label: string,
  fn: (screen: ScreenFormat) => void,
  expFlags: ExpFlag[],
): void {
  testScreen(label, fn, false, false, expFlags)
}
