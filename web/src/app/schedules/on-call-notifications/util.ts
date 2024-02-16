import { DateTime } from 'luxon'

import {
  OnCallNotificationRule,
  OnCallNotificationRuleInput,
  TargetType,
  WeekdayFilter,
} from '../../../schema'
import { allErrors, fieldErrors, nonFieldErrors } from '../../util/errutil'
import { weekdaySummary } from '../util'
import { ApolloError } from '@apollo/client'
import { CombinedError } from 'urql'

export type Value = {
  time: string | null
  weekdayFilter: WeekdayFilter
  type: TargetType
  targetID: string | null
}

export type RuleFieldError = {
  field: 'time' | 'weekdayFilter' | 'type' | 'slackChannelID' | 'slackUserGroup'
  message: string
}

export const EVERY_DAY = new Array(7).fill(true) as WeekdayFilter
export const NO_DAY = new Array(7).fill(false) as WeekdayFilter

export function formatLocalClockTime(dt: DateTime): string {
  const zoneStr = dt.toLocaleString(DateTime.TIME_SIMPLE)
  const localStr = dt.setZone('local').toLocaleString(DateTime.TIME_SIMPLE)

  if (zoneStr === localStr) {
    return ''
  }

  return `${localStr} ${dt.setZone('local').toFormat('ZZZZ')}`
}

export function formatClockTime(dt: DateTime): string {
  const zoneStr = dt.toLocaleString(DateTime.TIME_SIMPLE)
  const localStr = formatLocalClockTime(dt)
  if (!localStr) {
    return zoneStr
  }

  return `${zoneStr} (${localStr})`
}

// formatZoneTime will format a time string simply for the current locale
// if the time is different in local from zone, then it will additionally add
// the local time and offset as a suffix.
//
// example: 9:00 AM (4:00 AM CST)
function formatZoneTime(zone: string, time: string): string {
  if (!zone) {
    // fallback to just displaying the existing time according to locale
    return DateTime.fromFormat(time, 'HH:mm').toLocaleString(
      DateTime.TIME_SIMPLE,
    )
  }

  const dt = DateTime.fromFormat(time, 'HH:mm', { zone })

  return formatClockTime(dt)
}

export const onCallValueToRuleInput = (
  zone: string,
  v: Value,
): OnCallNotificationRuleInput => ({
  time: v.time
    ? DateTime.fromISO(v.time).setZone(zone).toFormat('HH:mm')
    : undefined,
  weekdayFilter: v.time ? v.weekdayFilter : undefined,
  target: { id: v.targetID || '', type: v.type },
})

export const onCallRuleToInput = (
  v: OnCallNotificationRule,
): OnCallNotificationRuleInput | null => {
  if (!v.target) return null

  return {
    time: v.time,
    id: v.id,
    weekdayFilter: v.weekdayFilter,
    target: { type: v.target.type, id: v.target.id },
  }
}

export function mapOnCallErrors(
  mErr?: ApolloError | CombinedError | null,
  ...qErr: Array<ApolloError | CombinedError | undefined>
): [Error[], RuleFieldError[]] {
  let dialogErrs: Error[] = []
  qErr.forEach((e) => (dialogErrs = dialogErrs.concat(allErrors(e))))
  if (!mErr) {
    return [dialogErrs, []]
  }

  dialogErrs = dialogErrs.concat(nonFieldErrors(mErr))
  const fieldErrs = fieldErrors(mErr)
    .map((e) => {
      switch (e.field) {
        case 'time':
        case 'weekdayFilter':
          return e
      }

      if (e.field === 'targetTypeSlackChannel') {
        return { ...e, field: 'slackChannelID' }
      }

      if (e.field === 'targetTypeSlackUserGroup') {
        return { ...e, field: 'slackUserGroup' }
      }

      dialogErrs.push(e)
      return null
    })
    .filter((e) => e !== null) as RuleFieldError[]

  return [dialogErrs, fieldErrs]
}

// SummaryInput will accept a partial OnCallNotificationRule, only depending on the fields
// that are used to generate the summary. This prevents typescript from complaining about
// missing fields when we only need a subset of the fields (like `target` which will be deprecated
// in favor of `dest`).
export type SummaryInput = Pick<
  OnCallNotificationRule,
  'time' | 'weekdayFilter'
>

export function onCallRuleSummary(zone: string, r?: SummaryInput): string {
  if (!r) return ''
  if (!r.time) return 'when on-call changes.'

  return `${weekdaySummary(r.weekdayFilter)} at ${formatZoneTime(zone, r.time)}`
}
