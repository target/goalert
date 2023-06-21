import { DateTime } from 'luxon'
import { ensureInterval } from './timeUtil'

describe('ensureInterval', () => {
  it('should preserve additional properties', () => {
    const val = {
      start: DateTime.utc().toISO(),
      end: DateTime.utc().plus({ minutes: 15 }).toISO(),
      extra: 1,
    }

    const newVal = ensureInterval(val, val)

    expect(newVal.extra).toEqual(1)
  })

  it('should not change start or end if it is a valid interval', () => {
    const n = DateTime.utc()
    const oldVal = {
      start: n.toISO(),
      end: n.plus({ minutes: 1 }).toISO(),
    }
    const newVal = {
      start: n.plus({ minutes: 2 }).toISO(),
      end: n.plus({ minutes: 3 }).toISO(),
    }

    const res = ensureInterval(oldVal, newVal)

    expect(res.start).toEqual(newVal.start)
    expect(res.end).toEqual(newVal.end)
  })

  it('should jump forward if start is past the new end time, and end is unchanged', () => {
    const n = DateTime.utc()
    const oldVal = {
      start: n.toISO(),
      end: n.plus({ minutes: 1 }).toISO(),
    }
    const newVal = {
      start: n.plus({ minutes: 2 }).toISO(),
      end: oldVal.end, // unchanged
    }

    const res = ensureInterval(oldVal, newVal)

    expect(res.start).toEqual(newVal.start)
    expect(res.end).toEqual(n.plus({ minutes: 3 }).toISO())
  })

  it('should jump backward if end is before the new start time', () => {
    const n = DateTime.utc()
    const oldVal = {
      start: n.toISO(),
      end: n.plus({ minutes: 1 }).toISO(),
    }
    const newVal = {
      start: oldVal.start, // unchanged
      end: n.minus({ minutes: 1 }).toISO(),
    }

    const res = ensureInterval(oldVal, newVal)

    expect(res.start).toEqual(n.minus({ minutes: 2 }).toISO())
    expect(res.end).toEqual(newVal.end)
  })
})
