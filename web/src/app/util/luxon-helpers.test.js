import { getLuxonStartOfWeek, getLuxonEndOfWeek } from './luxon-helpers'

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
  const result1 = getLuxonStartOfWeek()
  const result2 = getNativeStartOfWeek()
  expect(result1.toMillis()).toBe(result2.getTime())

  const davidsBirthday = new Date('November 26, 1997 03:24:00') // wednesday
  const startOfDavidsWeek = new Date('November 23, 1997 00:00:00')
  expect(getLuxonStartOfWeek(davidsBirthday).toMillis()).toBe(
    startOfDavidsWeek.getTime(),
  )

  const katiesBirthday = new Date('December 15, 1996 03:24:30') // sunday
  const startOfKatiesWeek = new Date('December 15, 1996 00:00:00')
  expect(getLuxonStartOfWeek(katiesBirthday).toMillis()).toBe(
    startOfKatiesWeek.getTime(),
  )

  const cooksBirthday = new Date('September 1, 2018 12:24:30') // saturday
  const startOfCooksWeek = new Date('August 26, 2018 00:00:00')
  expect(getLuxonStartOfWeek(cooksBirthday).toMillis()).toBe(
    startOfCooksWeek.getTime(),
  )

  const midnight = new Date('January 19, 2020 00:00:00') // 12am sunday
  expect(getLuxonStartOfWeek(midnight).toMillis()).toBe(midnight.getTime())
})

test('it should yield almost midnight on Saturday of the week of the date given', () => {
  // get end of today's week
  const result1 = getLuxonEndOfWeek()
  const result2 = getNativeEndOfWeek()
  expect(result1.toMillis()).toBe(result2.getTime())

  const davidsBirthday = new Date('November 26, 1997 03:24:00') // wednesday
  const endOfDavidsWeek = new Date('November 29, 1997 23:59:59:999')
  expect(getLuxonEndOfWeek(davidsBirthday).toMillis()).toBe(
    endOfDavidsWeek.getTime(),
  )

  const katiesBirthday = new Date('December 15, 1996 03:24:30') // sunday
  const endOfKatiesWeek = new Date('December 21, 1996 23:59:59:999')
  expect(getLuxonEndOfWeek(katiesBirthday).toMillis()).toBe(
    endOfKatiesWeek.getTime(),
  )

  const cooksBirthday = new Date('September 1, 2018 12:24:30') // saturday
  const endOfCooksWeek = new Date('September 1, 2018 23:59:59:999')
  expect(getLuxonEndOfWeek(cooksBirthday).toMillis()).toBe(
    endOfCooksWeek.getTime(),
  )

  const almostMidnight = new Date('January 18, 2020 23:59:59:999') // saturday
  expect(getLuxonEndOfWeek(almostMidnight).toMillis()).toBe(
    almostMidnight.getTime(),
  )
})
