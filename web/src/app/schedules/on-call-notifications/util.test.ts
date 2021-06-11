import {
  OnCallNotificationRule,
  OnCallNotificationRuleInput,
  WeekdayFilter,
} from '../../../schema'
import { getDayNames, mapDataToInput } from './util'

describe('mapDataToInput', () => {
  const check = (
    data: OnCallNotificationRule[],
    expected: OnCallNotificationRuleInput[],
  ): void => {
    const actual = mapDataToInput(data)
    expect(actual).toEqual(expected)
  }

  it('should transform scheduled notification', () => {
    check(
      [
        {
          id: '4ca5fe3c-b970-4e47-a218-c7deb15bcc94:0',
          target: {
            id: 'H4M8URG3RRR',
            type: 'slackChannel',
            name: '#hamburger',
          },
          time: '09:00',
          weekdayFilter: [false, true, true, false, true, true, false],
        },
      ],
      [
        {
          id: '4ca5fe3c-b970-4e47-a218-c7deb15bcc94:0',
          target: {
            id: 'H4M8URG3RRR',
            type: 'slackChannel',
          },
          time: '09:00',
          weekdayFilter: [false, true, true, false, true, true, false],
        },
      ],
    )
  })

  it('should transform on-change notification', () => {
    check(
      [
        {
          id: '5ca5fe3c-b970-4e47-a218-c7deb15bcc94:0',
          target: {
            id: 'FR3NCHFR135',
            type: 'slackChannel',
            name: '#frenchfries',
          },
        },
      ],
      [
        {
          id: '5ca5fe3c-b970-4e47-a218-c7deb15bcc94:0',
          target: {
            id: 'FR3NCHFR135',
            type: 'slackChannel',
          },
        },
      ],
    )
  })

  it('should transform multiple notification types', () => {
    check(
      [
        {
          id: '4ca5fe3c-b970-4e47-a218-c7deb15bcc94:0',
          target: {
            id: 'H4M8URG3RRR',
            type: 'slackChannel',
            name: '#hamburger',
          },
          time: '09:00',
          weekdayFilter: [false, true, true, false, true, true, false],
        },
        {
          id: '5ca5fe3c-b970-4e47-a218-c7deb15bcc94:0',
          target: {
            id: 'FR3NCHFR135',
            type: 'slackChannel',
            name: '#frenchfries',
          },
        },
        {
          id: '6ca5fe3c-b970-4e47-a218-c7deb15bcc94:0',
          target: {
            id: 'V3G4NT4C055',
            type: 'slackChannel',
            name: '#vegantacos',
          },
          time: '00:00',
          weekdayFilter: [false, false, true, false, false, false, false],
        },
      ],
      [
        {
          id: '4ca5fe3c-b970-4e47-a218-c7deb15bcc94:0',
          target: {
            id: 'H4M8URG3RRR',
            type: 'slackChannel',
          },
          time: '09:00',
          weekdayFilter: [false, true, true, false, true, true, false],
        },
        {
          id: '5ca5fe3c-b970-4e47-a218-c7deb15bcc94:0',
          target: {
            id: 'FR3NCHFR135',
            type: 'slackChannel',
          },
        },
        {
          id: '6ca5fe3c-b970-4e47-a218-c7deb15bcc94:0',
          target: {
            id: 'V3G4NT4C055',
            type: 'slackChannel',
          },
          time: '00:00',
          weekdayFilter: [false, false, true, false, false, false, false],
        },
      ],
    )
  })
})

describe('getDayNames', () => {
  const check = (filter: WeekdayFilter, expected: string): void => {
    expect(getDayNames(filter)).toEqual(expected)
  }

  check([true, true, true, true, true, true, true], 'every day')
  check([false, true, true, true, true, true, false], 'weekdays')
  check(
    [false, false, true, true, true, true, false],
    'Tuesdays, Wednesdays, Thursdays, and Fridays',
  )
  check([true, false, false, false, false, false, false], 'Sundays')
  check([false, false, false, false, false, false, true], 'Saturdays')
  check(
    [false, false, false, false, false, true, true],
    'Fridays and Saturdays',
  )
})
