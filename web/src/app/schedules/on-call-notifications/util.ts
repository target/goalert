import { LocalZone, Zone } from 'luxon'
import { OnCallNotificationRuleInput, WeekdayFilter } from '../../../schema'
import { days, isoToGQLClockTime } from '../util'

export interface Rule {
  id: string
  target: {
    id: string
    type: string
    name?: string
  }
  time?: string
  weekdayFilter?: WeekdayFilter
}

export function mapDataToInput(
  rules: Array<Rule> = [],
  // TODO remove default
  scheduleTimeZone: Zone = LocalZone.instance,
): Array<OnCallNotificationRuleInput> {
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
  }) as Array<OnCallNotificationRuleInput>
}

export function getDayNames(filter: WeekdayFilter): string {
  const names = days.filter((name, i) => filter[i]).map((day) => day + 's')
  const lastDay = names.length > 1 ? names.pop() : ''
  return names.join(', ') + (lastDay && ' and ' + lastDay)
}
