import {
  OnCallNotificationRule,
  OnCallNotificationRuleInput,
} from '../../../schema'
import { mapDataToInput } from './util'

describe('mapDataToInput', () => {
  const check = (
    data: OnCallNotificationRule[],
    input: OnCallNotificationRuleInput[],
  ): void => {
    const actual = mapDataToInput(data)
    const expected = input
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
