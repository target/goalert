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

export function getDayNames(filter: WeekdayFilter): string {
  const everyday = [true, true, true, true, true, true, true]
  const isEverday = filter.every((val, i) => val === everyday[i])
  if (isEverday) return 'every day'

  const weekdays = [false, true, true, true, true, true, false]
  const isWeekdays = filter.every((val, i) => val === weekdays[i])
  if (isWeekdays) return 'weekdays'

  const names = days.filter((name, i) => filter[i]).map((day) => day + 's')
  const lastDay = names.length > 1 ? names.pop() : ''
  return names.join(', ') + (lastDay && ' and ' + lastDay)
}
