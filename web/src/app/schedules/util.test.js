import { mapOverrideUserError, alignWeekdayFilter, mapRuleTZ } from './util'
import _ from 'lodash'

const fromBin = (f) => f.split('').map((f) => f === '1')

describe('mapRuleTZ', () => {
  const check = (rule, fromTZ, toTZ, expected) => {
    rule.weekdayFilter = fromBin(rule.f)
    expected.weekdayFilter = fromBin(expected.f)

    expect(mapRuleTZ(fromTZ, toTZ, _.omit(rule, 'f'))).toEqual(
      _.omit(expected, 'f'),
    )
  }
  it('should not change same TZ', () => {
    check({ start: '00:00', end: '00:00', f: '1000000' }, 'UTC', 'UTC', {
      start: '00:00',
      end: '00:00',
      f: '1000000',
    })
  })

  it('should map across days, and back', () => {
    check({ start: '00:00', end: '00:00', f: '1000000' }, 'UTC', 'UTC-6', {
      start: '18:00',
      end: '18:00',
      f: '0000001',
    })

    check(
      {
        start: '18:00',
        end: '18:00',
        f: '0000001',
      },
      'UTC-6',
      'UTC',
      { start: '00:00', end: '00:00', f: '1000000' },
    )
  })
})

describe('alignWeekdayFilter', () => {
  const check = (input, n, expected) =>
    expect(alignWeekdayFilter(n, fromBin(input))).toEqual(fromBin(expected))

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

  const add = { ...data, addUser: { id: 'foo', name: 'bob' } }
  const remove = { ...data, removeUser: { id: 'bar', name: 'ben' } }
  const replace = { ...add, ...remove }

  const check = (override, value, errs) =>
    expect(mapOverrideUserError(override, value, zone)).toEqual(errs)

  it('should generate proper error messages', () => {
    check(add, { addUserID: 'foo' }, [
      {
        field: 'addUserID',
        message: 'Already added from ' + timeStr,
      },
    ])

    check(replace, { addUserID: 'bar' }, [
      {
        field: 'addUserID',
        message: 'Already replaced by bob from ' + timeStr,
      },
    ])
    check(replace, { addUserID: 'foo' }, [
      {
        field: 'addUserID',
        message: 'Already replacing ben from ' + timeStr,
      },
    ])
    check(remove, { addUserID: 'bar' }, [
      {
        field: 'addUserID',
        message: 'Already removed from ' + timeStr,
      },
    ])

    check(add, { removeUserID: 'foo' }, [
      {
        field: 'removeUserID',
        message: 'Already added from ' + timeStr,
      },
    ])
  })
})
