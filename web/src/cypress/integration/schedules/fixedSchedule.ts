import { Chance } from 'chance'
import { testScreen } from '../../support'
import { Schedule } from '../../../schema'

const c = new Chance()

function testFixedSchedule(screen: ScreenFormat): void {
  beforeEach(() => {
    cy.createSchedule().then((sched: Schedule) => {
      cy.visit('/schedules/' + sched.id)
      cy.get('[data-cy="new-fixed-sched"]').click()
    })
  })

  // todo: need to be reasonable with random dates
  // constrain to current month? random lengths from 1d to a month?

  it.only('should create a fixed schedule', () => {
    // fill out step 1 start and end times
    cy.dialogForm({
      start: '', // random date
      end: '', // random date after start
    })

    // go to step 2
    cy.get('[data-cy="loading-button"]').contains('Next').click()

    // add shift
    cy.dialogForm({
      start: '', // random date
      end: '', // random date after start
      userID: '', // not sure how this works with select drop downs
    })

    // click add shift button

    // verify shift shows up on right

    // click submit

    // check fixed sched length in calendar

    // check new shift in calendar
  })

  it('should create a fixed schedule overlapping existing shifts', () => {})
  it('should edit a fixed schedule', () => {})
  it('should delete a fixed schedule', () => {})

  it('should toggle timezone', () => {})
  it('should toggle duration field', () => {})
  it('should delete a shift in step 2', () => {})
  it('should go back and forth between steps', () => {})
  it('should cancel and close form', () => {})
}

testScreen('Fixed Schedule', testFixedSchedule)
