import { randInterval, testScreen } from '../support'
import { Schedule, User } from '../../schema'
import { DateTime } from 'luxon'

const dtFmt = "yyyy-MM-dd'T'HH:mm"
const dialog = '[role=dialog] #dialog-form'

function testTemporarySchedule(screen: string): void {
  if (screen !== 'widescreen') return

  let schedule: Schedule
  let manualAddUser: User
  let graphQLAddUser: User
  beforeEach(() => {
    cy.fixture('users').then((u) => {
      manualAddUser = u[0]
      graphQLAddUser = u[1]

      cy.createSchedule({ timeZone: 'Europe/Berlin' }).then((s: Schedule) => {
        schedule = s
        cy.visit('/schedules/' + s.id)
      })
    })
  })

  const schedTZ = (t: DateTime): string =>
    t.setZone(schedule.timeZone).toFormat(dtFmt)

  it('should toggle duration field', () => {
    const defaultDurationHrs = 8
    cy.get('[data-cy="new-override"]').click()
    cy.dialogTitle('Choose')
    cy.dialogForm({ variant: 'temp' })
    cy.dialogClick('Next')

    cy.get('[data-cy="add-shift-expander"]').click()

    // check default state of duration
    cy.get(dialog)
      .find('input[name="shift-end"]')
      .should('have.value', defaultDurationHrs)
    cy.get(dialog).find('[data-cy="toggle-duration-off"]').click()

    cy.get('input[name="shift-start"]')
      .invoke('val')
      .then((shiftStart) => {
        cy.get('input[name="shift-end"]')
          .invoke('val')
          .should((shiftEnd) => {
            expect(DateTime.fromISO(shiftEnd as string).toISO()).to.eq(
              DateTime.fromISO(shiftStart as string)
                .plus({ hour: defaultDurationHrs })
                .toISO(),
            )
          })
      })
  })

  it('should cancel and close form', () => {
    cy.get('[role="dialog"]').should('not.exist')
    cy.get('[data-cy="new-override"]').click()
    cy.dialogTitle('Choose')
    cy.dialogForm({ variant: 'temp' })
    cy.dialogClick('Next')
    cy.get('[role="dialog"]').should('be.visible')
    cy.dialogFinish('Cancel')
    cy.get('[role="dialog"]').should('not.exist')
  })

  it('should create a temporary schedule', () => {
    const { start, end } = randInterval()
    cy.get('[data-cy="new-override"]').click()
    cy.dialogTitle('Choose')
    cy.dialogForm({ variant: 'temp' })
    cy.dialogClick('Next')
    cy.dialogForm({ start: schedTZ(start), end: schedTZ(end) })
    cy.get('[data-cy="add-shift-expander"]').click()
    cy.get('[data-cy="no-coverage-checkbox"]').should('not.exist')
    cy.get('[data-cy="shifts-list"]').should('not.contain', manualAddUser.name)
    cy.dialogForm({ userID: manualAddUser.name })
    cy.get('button[data-cy="add-shift"]').click()
    cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)

    cy.dialogClick('Submit')
    cy.get('[data-cy="no-coverage-checkbox"]')
      .should('be.visible')
      .find('input[name="allowCoverageGaps"]')
      .check()
    cy.dialogFinish('Retry')

    cy.visit('/schedules/' + schedule.id + '?start=' + start.toISO())
    cy.get('div').contains('Temporary Schedule').click()
    cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
    cy.get('button[data-cy="edit-temp-sched"]').should('be.visible')
    cy.get('button[data-cy="delete-temp-sched"]').should('be.visible')
    cy.get('body').trigger('keydown', { key: 'Escape' })
    cy.get('div').contains(manualAddUser.name).click()
    cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
  })

  // seems buggy when shift list has overflow
  it('should edit a temporary schedule', () => {
    const now = DateTime.utc()

    cy.createTemporarySchedule({
      scheduleID: schedule.id,
      start: now.toISO(),
      shifts: [{ userID: graphQLAddUser.id }],
    }).then(() => {
      cy.reload()
      cy.get('div').contains('Temporary Schedule').click()
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
      cy.get('button[data-cy="edit-temp-sched"]').click()
      cy.get('[data-cy="shifts-list"]').should('contain', graphQLAddUser.name)
      cy.get('[data-cy="shifts-list"] li [aria-label="delete shift"]').click({
        force: true,
      }) // delete
      cy.get('[data-cy="shifts-list"]').should(
        'not.contain',
        graphQLAddUser.name,
      )

      cy.get('[data-cy="add-shift-expander"]').click()
      cy.dialogForm({
        userID: manualAddUser.name,
        'shift-start': schedTZ(now.plus({ hour: 1 })),
      })
      cy.get('[data-cy="shifts-list"]').should(
        'not.contain',
        manualAddUser.name,
      )
      cy.get('button[data-cy="add-shift"]').click()
      cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)
      cy.dialogClick('Submit')
      cy.get('[data-cy="no-coverage-checkbox"]')
        .should('be.visible')
        .find('input[name="allowCoverageGaps"]')
        .check()
      cy.dialogFinish('Retry')
      cy.reload() // ensure calendar update
      cy.get('div').contains(manualAddUser.name).click()
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
    })
  })

  it('should delete a temporary schedule', () => {
    cy.createTemporarySchedule({
      start: DateTime.utc().plus({ hour: 1 }).toISO(),
      scheduleID: schedule.id,
    }).then(() => {
      cy.reload()
      cy.get('div').contains('Temporary Schedule').click()
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
      cy.get('button[data-cy="delete-temp-sched"]').click()
      cy.dialogFinish('Confirm')
      cy.get('div').contains('Temporary Schedule').should('not.exist')
    })
  })

  it('should be able to add multiple shifts', () => {
    const ivl = randInterval()
    cy.get('[data-cy="new-override"]').click()
    cy.dialogTitle('Choose')
    cy.dialogForm({ variant: 'temp' })
    cy.dialogClick('Next')
    cy.dialogForm({
      start: schedTZ(ivl.start),
      end: schedTZ(ivl.end),
    })
    cy.get('[data-cy="add-shift-expander"]').click()

    // add first shift
    cy.dialogForm({
      userID: manualAddUser.name,
      'shift-start': schedTZ(ivl.start),
      'shift-end': (ivl.toDuration().as('hours') / 3).toFixed(2),
    })
    cy.get('[data-cy="shifts-list"]').should('not.contain', manualAddUser.name)
    cy.get('button[data-cy="add-shift"]').click()
    cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)

    // add second shift
    cy.dialogForm({
      userID: graphQLAddUser.name, // gql not in this test, safe to use here
    })
    cy.get('[data-cy="shifts-list"]').should('not.contain', graphQLAddUser.name)
    cy.get('button[data-cy="add-shift"]').click()

    cy.get('[data-cy="shifts-list"]').should('contain', graphQLAddUser.name)
    cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)
  })
}

testScreen('temporary Schedule', testTemporarySchedule)
