import { calcNewActiveIndex, handoffSummary, reorderList } from './util'
import { Settings } from 'luxon'

describe('calcNewActiveIndex', () => {
  let oldZone = ''
  beforeAll(() => {
    // 01/02 03:04:05PM '06 -0700
    Settings.now = () => 1136239445
    oldZone = Settings.defaultZoneName
    Settings.defaultZoneName = 'UTC'
  })
  afterAll(() => {
    Settings.now = () => Date.now()
    Settings.defaultZone = oldZone
  })
  const check = (aIdx, oldIdx, newIdx, exp) => {
    expect(calcNewActiveIndex(aIdx, oldIdx, newIdx)).toBe(exp)
  }
  test('should return -1 when no change', () => {
    check(0, 1, 2, -1)
    check(0, 2, 1, -1)
    check(3, 1, 2, -1)
    check(3, 2, 1, -1)
    check(0, 0, 0, -1)
  })
  test('should return newIndex when active user is being dragged', () => {
    check(1, 1, 2, 2)
    check(2, 2, 1, 1)
  })
  test('should return newIndex +1 ', () => {
    check(0, 2, 0, 1)
    check(1, 2, 0, 2)
  })
  test('should return newIndex -1', () => {
    check(1, 0, 2, 0)
    check(1, 0, 1, 0)
  })
})

describe('handoffSummary', () => {
  const check = (rotation, exp) => {
    expect(handoffSummary(rotation)).toBe(exp)
  }

  test('should be as per hourly rotation', () => {
    check(
      {
        shiftLength: 1,
        start: '2018-07-25T02:22:33Z',
        timeZone: 'UTC',
        type: 'hourly',
      },
      'First hand off time at 2:22 AM UTC, hands off every hour.',
    )

    check(
      {
        shiftLength: 1,
        start: '2017-07-14T06:32:33Z',
        timeZone: 'Asia/Kolkata',
        type: 'hourly',
      },
      'First hand off time at 12:02 PM Asia/Kolkata (6:32 AM local), hands off every hour.',
    )
  })

  test('should be as per daily rotation', () => {
    check(
      {
        shiftLength: 2,
        start: '2018-02-25T09:10:22Z',
        timeZone: 'America/Cancun',
        type: 'daily',
      },
      'Hands off every 2 days at 4:10 AM America/Cancun (9:10 AM local).',
    )

    check(
      {
        shiftLength: 1,
        start: '2017-07-14T06:32:33Z',
        timeZone: 'UTC',
        type: 'daily',
      },
      'Hands off daily at 6:32 AM UTC.',
    )
  })

  test('should be as per weekly rotation', () => {
    check(
      {
        shiftLength: 2,
        start: '2018-02-25T09:10:22Z',
        timeZone: 'UTC',
        type: 'weekly',
      },
      'Hands off every 2 weeks on Sunday at 9:10 AM UTC.',
    )

    check(
      {
        shiftLength: 2,
        start: '2017-06-26T06:50:11Z',
        timeZone: 'Asia/Kolkata',
        type: 'weekly',
      },
      'Hands off every 2 weeks on Monday at 12:20 PM Asia/Kolkata (Monday at 6:50 AM local time).',
    )
  })
})

describe('reorderList', () => {
  const check = (users, oldIdx, newIdx, exp) => {
    expect(reorderList(users, oldIdx, newIdx)).toEqual(exp)
  }

  test('should return reordered user list', () => {
    check(['aaa', 'bbb', 'ccc'], 0, 0, ['aaa', 'bbb', 'ccc'])
    check(['aaa', 'bbb', 'ccc'], 0, 1, ['bbb', 'aaa', 'ccc'])
    check(['aaa', 'bbb', 'ccc'], 0, 2, ['bbb', 'ccc', 'aaa'])
    check(['aaa', 'bbb', 'ccc'], 1, 0, ['bbb', 'aaa', 'ccc'])
    check(['aaa', 'bbb', 'ccc'], 1, 1, ['aaa', 'bbb', 'ccc'])
    check(['aaa', 'bbb', 'ccc'], 1, 2, ['aaa', 'ccc', 'bbb'])
    check(['aaa', 'bbb', 'ccc'], 2, 0, ['ccc', 'aaa', 'bbb'])
    check(['aaa', 'bbb', 'ccc'], 2, 1, ['aaa', 'ccc', 'bbb'])
    check(['aaa', 'bbb', 'ccc'], 2, 2, ['aaa', 'bbb', 'ccc'])

    check(['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'], 6, 0, [
      'g',
      'a',
      'b',
      'c',
      'd',
      'e',
      'f',
      'h',
    ])

    check(['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'], 3, 7, [
      'a',
      'b',
      'c',
      'e',
      'f',
      'g',
      'h',
      'd',
    ])

    check(['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'], 3, 0, [
      'd',
      'a',
      'b',
      'c',
      'e',
      'f',
      'g',
      'h',
    ])
  })
})
