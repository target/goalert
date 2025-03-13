import { UserOverride, WeekdayFilter } from '../../schema'
import { FieldError } from '../util/errutil'
import { mapOverrideUserError, alignWeekdayFilter } from './util'
import _ from 'lodash'

const fromBin = (f: string): boolean[] => f.split('').map((f) => f === '1')

describe('alignWeekdayFilter', () => {
  const check = (input: string, n: number, expected: string) =>
    expect(alignWeekdayFilter(n, fromBin(input) as WeekdayFilter)).toEqual(fromBin(expected))

  it('should leave aligned filters alone', () => {
    check('1010101', 7, '1010101')
    check('1010001', 7, '1010001')
    check('1111111', 7, '1111111')
  })

  it('should align differences', () => {
    // sunday becomes sat
    check('1000000', 6, '0000001')
    check('0010000', 6, '0100000')
    // sunday becomes mon
    check('1000000', 1, '0100000')
    check('0010000', 1, '0001000')
  })
})

describe('mapOverrideUserError', () => {
  const data = {
    start: '2019-01-02T20:33:10.363Z',
    end: '2019-01-02T21:33:10.363Z',
  }
  const timeStr = 'Jan 2, 2019, 2:33 PM to 3:33 PM'
  const zone = 'America/Chicago'

  const add = { ...data, addUser: { id: 'foo', name: 'bob' } } as UserOverride
  const remove = { ...data, removeUser: { id: 'bar', name: 'ben' } } as UserOverride
  const replace = { ...add, ...remove } as UserOverride

  const check = (override: UserOverride, value: UserOverride, errs: FieldError[]) =>
    expect(mapOverrideUserError(override, value, zone)).toEqual(errs)

  it('should generate proper error messages', () => {
    check(add, { addUserID: 'foo' } as UserOverride, [
      {
        field: 'addUserID',
        message: 'Already added from ' + timeStr,
      },
    ] as FieldError[])

    check(replace, { addUserID: 'bar' } as UserOverride, [
      {
        field: 'addUserID',
        message: 'Already replaced by bob from ' + timeStr,
      },
    ] as FieldError[])
    check(replace, { addUserID: 'foo' } as UserOverride, [
      {
        field: 'addUserID',
        message: 'Already replacing ben from ' + timeStr,
      },
    ] as FieldError[])
    check(remove, { addUserID: 'bar' } as UserOverride, [
      {
        field: 'addUserID',
        message: 'Already removed from ' + timeStr,
      },
    ] as FieldError[])

    check(add, { removeUserID: 'foo' } as UserOverride, [
      {
        field: 'removeUserID',
        message: 'Already added from ' + timeStr,
      },
    ] as FieldError[])
  })
})
