import { DateTime, Interval } from 'luxon'
import { Chance } from 'chance'
import {
  getStartOfWeek,
  getEndOfWeek,
  splitShift,
  getNextWeekday,
} from './luxon-helpers'

const getNativeStartOfWeek = (dt = new Date()): Date => {
  const weekdayIndex = dt.getDay() // Sun - Sat : 0 - 6
  const sunday = new Date(new Date(dt).setDate(dt.getDate() - weekdayIndex))
  return new Date(new Date(sunday).setHours(0, 0, 0, 0))
}

const getNativeEndOfWeek = (dt = new Date()): Date => {
  const weekdayIndex = dt.getDay() // Sun - Sat : 0 - 6
  const saturday = new Date(
    new Date(dt).setDate(dt.getDate() + (6 - weekdayIndex)),
  )
  return new Date(new Date(saturday).setHours(23, 59, 59, 999))
}

test('it should yield 12am on Sunday of the week of the date given', () => {
  // get beginning of today's week
  const result1 = getStartOfWeek()
  const result2 = getNativeStartOfWeek()
  expect(result1.toMillis()).toBe(result2.getTime())

  const randomWednesday = DateTime.fromJSDate(
    new Date('November 26, 1997 03:24:00'),
  )
  const startOfRandomWednesdaysWeek = new Date('November 23, 1997 00:00:00')
  expect(getStartOfWeek(randomWednesday).toMillis()).toBe(
    startOfRandomWednesdaysWeek.getTime(),
  )

  const randomSunday = DateTime.fromJSDate(
    new Date('December 15, 1996 03:24:30'),
  )
  const startOfRandomSundaysWeek = new Date('December 15, 1996 00:00:00')
  expect(getStartOfWeek(randomSunday).toMillis()).toBe(
    startOfRandomSundaysWeek.getTime(),
  )

  const randomSaturday = DateTime.fromJSDate(
    new Date('September 1, 2018 12:24:30'),
  )
  const startOfRandomSaturdaysWeek = new Date('August 26, 2018 00:00:00')
  expect(getStartOfWeek(randomSaturday).toMillis()).toBe(
    startOfRandomSaturdaysWeek.getTime(),
  )

  const midnight = new Date('January 19, 2020 00:00:00') // 12am sunday
  expect(getStartOfWeek(DateTime.fromJSDate(midnight)).toMillis()).toBe(
    midnight.getTime(),
  )
})

test('it should yield almost midnight on Saturday of the week of the date given', () => {
  // get end of today's week
  const result1 = getEndOfWeek()
  const result2 = getNativeEndOfWeek()
  expect(result1.toMillis()).toBe(result2.getTime())

  const randomWednesday = DateTime.fromJSDate(
    new Date('November 26, 1997 03:24:00'),
  )
  const endOfRandomWednesdaysWeek = new Date('November 29, 1997 23:59:59:999')
  expect(getEndOfWeek(randomWednesday).toMillis()).toBe(
    endOfRandomWednesdaysWeek.getTime(),
  )

  const randomSunday = DateTime.fromJSDate(
    new Date('December 15, 1996 03:24:30'),
  )
  const endOfRandomSundaysWeek = new Date('December 21, 1996 23:59:59:999')
  expect(getEndOfWeek(randomSunday).toMillis()).toBe(
    endOfRandomSundaysWeek.getTime(),
  )

  const randomSaturday = DateTime.fromJSDate(
    new Date('September 1, 2018 12:24:30'),
  )
  const endOfRandomSaturdaysWeek = new Date('September 1, 2018 23:59:59:999')
  expect(getEndOfWeek(randomSaturday).toMillis()).toBe(
    endOfRandomSaturdaysWeek.getTime(),
  )

  const almostMidnight = new Date('January 18, 2020 23:59:59:999') // saturday
  expect(getEndOfWeek(DateTime.fromJSDate(almostMidnight)).toMillis()).toBe(
    almostMidnight.getTime(),
  )
})

describe('getNextWeekday', () => {
  const chicago = 'America/Chicago'
  const singapore = 'Asia/Singapore'

  it('get next monday from sunday', () => {
    const result = getNextWeekday(
      1, // monday
      DateTime.fromObject({ year: 2021, month: 8, day: 8 }), // sunday
      chicago,
    )
    expect(result).toEqual(
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 9,
        },
        { zone: chicago },
      ),
    )
  })

  it('get next tues from sunday', () => {
    const result = getNextWeekday(
      2, // tues
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 8,
        },
        { zone: chicago },
      ), // sunday
      chicago,
    )
    expect(result).toEqual(
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 10,
        },
        { zone: chicago },
      ),
    )
  })

  it('get next sat from sunday', () => {
    const result = getNextWeekday(
      6, // saturday
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 8,
        },
        { zone: chicago },
      ), // sunday
      chicago,
    )
    expect(result).toEqual(
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 14,
        },
        { zone: chicago },
      ),
    )
  })

  it('get next sun from sunday', () => {
    const result = getNextWeekday(
      7,
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 8,
        },
        { zone: chicago },
      ),
      chicago,
    )
    expect(result).toEqual(
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 15,
        },
        { zone: chicago },
      ),
    )
  })

  it('get next friday from saturday', () => {
    const result = getNextWeekday(
      5,
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 7,
        },
        { zone: chicago },
      ),
      chicago,
    )
    expect(result).toEqual(
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 13,
        },
        { zone: chicago },
      ),
    )
  })

  it('get next sunday from tues', () => {
    const result = getNextWeekday(
      7,
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 10,
        },
        { zone: chicago },
      ),
      chicago,
    )
    expect(result).toEqual(
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 15,
        },
        { zone: chicago },
      ),
    )
  })

  it('get next sunday in asia from late-night saturday in america', () => {
    const result = getNextWeekday(
      7,
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 14,
          hour: 23,
        },
        { zone: chicago },
      ),
      singapore,
    )
    expect(result).toEqual(
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 22,
        },
        {
          zone: singapore,
        },
      ),
    )
  })

  it('get next sunday in america from saturday in asia', () => {
    const result = getNextWeekday(
      7,
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 14,
        },
        {
          zone: singapore,
        },
      ),
      chicago,
    )
    expect(result).toEqual(
      DateTime.fromObject(
        {
          year: 2021,
          month: 8,
          day: 15,
        },
        { zone: chicago },
      ),
    )
  })
})

describe('splitShift', () => {
  function rand(): DateTime {
    const c = new Chance()
    return DateTime.fromJSDate(c.date())
  }

  it('should handle interval that does not span midnight', () => {
    const hour = Interval.fromDateTimes(
      DateTime.fromObject({ hour: 1 }),
      DateTime.fromObject({ hour: 2 }),
    )
    const result = splitShift(hour)
    expect(result.length).toEqual(1)
    expect(result[0]).toEqual(hour)
  })

  it('should handle interval that begins at midnight, ends same day', () => {
    const startOfDay = Interval.fromDateTimes(
      DateTime.fromObject({ hour: 0 }).startOf('day'),
      DateTime.fromObject({ hour: 2 }),
    )
    const result = splitShift(startOfDay)
    expect(result.length).toEqual(1)
    expect(result[0]).toEqual(startOfDay)
  })

  it('should handle interval that begins at midnight, ends at midnight', () => {
    const start = rand().startOf('day')
    const end = start.plus({ day: 1 }).startOf('day')
    const inv = Interval.fromDateTimes(start, end)
    const result = splitShift(inv)

    expect(result.length).toEqual(1)
    expect(result[0]).toEqual(inv)
  })

  it('should handle interval that begins midday, ends at midnight', () => {
    const start = rand().set({ hour: 4 })
    const end = start.plus({ day: 1 }).startOf('day')
    const inv = Interval.fromDateTimes(start, end)
    const result = splitShift(inv)

    expect(result.length).toEqual(1)
    expect(result[0]).toEqual(inv)
  })

  it('should handle interval that begins at midnight, ends next day', () => {
    const start = rand().startOf('day')
    const end = start.plus({ day: 1, hour: 4 })
    const inv = Interval.fromDateTimes(start, end)
    const result = splitShift(inv)

    expect(result.length).toEqual(2)
    expect(result[0]).toEqual(Interval.fromDateTimes(start, end.startOf('day')))
    expect(result[1]).toEqual(Interval.fromDateTimes(end.startOf('day'), end))
  })

  it('should handle interval that spans multiple days', () => {
    const start = rand().set({ hour: 4 })
    const end = start.plus({ day: 3 })
    const inv = Interval.fromDateTimes(start, end)
    const result = splitShift(inv)

    expect(result.length).toEqual(4)
    expect(result[0]).toEqual(
      Interval.fromDateTimes(start, start.plus({ day: 1 }).startOf('day')),
    )
    expect(result[1]).toEqual(
      Interval.fromDateTimes(
        start.plus({ day: 1 }).startOf('day'),
        start.plus({ day: 2 }).startOf('day'),
      ),
    )
    expect(result[2]).toEqual(
      Interval.fromDateTimes(
        start.plus({ day: 2 }).startOf('day'),
        start.plus({ day: 3 }).startOf('day'),
      ),
    )
    expect(result[3]).toEqual(
      Interval.fromDateTimes(start.plus({ day: 3 }).startOf('day'), end),
    )
  })
})
