import { DateTime } from 'luxon'
import {
  OnCallNotificationRule,
  OnCallNotificationRuleInput,
  WeekdayFilter,
} from '../../../schema'
import { days, isoToGQLClockTime } from '../util'

// type aliases for convenience
export type Rule = OnCallNotificationRule
export type RuleInput = OnCallNotificationRuleInput

export function mapDataToInput(
  rules: Array<Rule> = [],
  scheduleTimeZone: string,
): Array<RuleInput> {
  return rules.map((r: Rule) => {
    const result: Rule = {
      id: r.id,
      target: {
        id: r.target.id,
        type: r.target.type,
      },
    }

    if (r.time) {
      result.time = isoToGQLClockTime(r.time, scheduleTimeZone)
    }
    if (r.weekdayFilter) {
      result.weekdayFilter = r.weekdayFilter
    }
    return result
  }) as Array<RuleInput>
}

export function weekdayFilterString(filter: WeekdayFilter): string {
  const names = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const state = filter.map((v) => (v ? '1' : '0')).join('')
  switch (state) {
    case '0000000':
      return 'never'
    case '1111111':
      return 'every day'
    case '1000001':
      return 'weekends'
    case '0111110':
      return 'M-F'
    case '11111110':
      return 'M-F and Sun'
    case '0111111':
      return 'M-F and Sat'
  }

  return filter
    .map((v, idx) => (v ? names[idx] : null))
    .filter((v) => v)
    .join(',')
}

export function getDayNames(filter: WeekdayFilter): string {
  const everyday = [true, true, true, true, true, true, true]
  const isEverday = filter.every((val, i) => val === everyday[i])
  if (isEverday) return 'every day'

  const weekdays = [false, true, true, true, true, true, false]
  const isWeekdays = filter.every((val, i) => val === weekdays[i])
  if (isWeekdays) return 'weekdays'

  const names = days.filter((name, i) => filter[i]).map((day) => day + 's')
  const lastDay = names.length > 1 ? names.pop() : ''
  const oxford = names.length > 1 ? ',' : ''
  return names.join(', ') + (lastDay && `${oxford} and ` + lastDay)
}

export function getRuleSummary(
  rule: Rule,
  scheduleZone: string,
  displayZone: string,
): string {
  if (rule.time && rule.weekdayFilter) {
    const timeStr = DateTime.fromFormat(rule.time, 'HH:mm', {
      zone: scheduleZone,
    })
      .setZone(displayZone)
      .toFormat('h:mm a ZZZZ')

    return `Notifies ${getDayNames(rule.weekdayFilter)} at ${timeStr}`
  }

  return 'Notifies when on-call hands off'
}
