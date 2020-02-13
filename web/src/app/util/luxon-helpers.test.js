import { getStartOfWeek, getEndOfWeek } from './luxon-helpers'
import { DateTime } from 'luxon'

const getNativeStartOfWeek = (dt = new Date()) => {
  const weekdayIndex = dt.getDay() // Sun - Sat : 0 - 6
  const sunday = new Date(new Date(dt).setDate(dt.getDate() - weekdayIndex))
  return new Date(new Date(sunday).setHours(0, 0, 0, 0))
}

const getNativeEndOfWeek = (dt = new Date()) => {
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
