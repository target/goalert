import { calcNewActiveIndex, reorderList } from './util'
import { Settings, Zone } from 'luxon'

let oldZone: Zone
beforeAll(() => {
  // 01/02 03:04:05PM '06 -0700
  Settings.now = () => 1136239445
  oldZone = Settings.defaultZone as Zone // getting defaultZone always returns a Zone object
  Settings.defaultZone = 'UTC'
})
afterAll(() => {
  Settings.now = () => Date.now()
  Settings.defaultZone = oldZone
})

describe('calcNewActiveIndex', () => {
  const check = (
    aIdx: number,
    oldIdx: number,
    newIdx: number,
    exp: number,
  ): void => {
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

describe('reorderList', () => {
  const check = (
    users: string[],
    oldIdx: number,
    newIdx: number,
    exp: string[],
  ): void => {
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
