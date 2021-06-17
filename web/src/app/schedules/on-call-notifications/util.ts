import { ApolloError } from '@apollo/client'
import { DateTime } from 'luxon'

import {
  OnCallNotificationRule,
  OnCallNotificationRuleInput,
  WeekdayFilter,
} from '../../../schema'
import { allErrors, fieldErrors, nonFieldErrors } from '../../util/errutil'
import { weekdaySummary } from '../util'

export type Value = {
  slackChannelID: string | null
  time: string | null
  weekdayFilter: WeekdayFilter
}

export type RuleFieldError = {
  field: 'time' | 'weekdayFilter' | 'slackChannelID'
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

function formatTime(zone: string, time: string): string {
  if (!zone) {
    // fallback to just displaying the existing time according to locale
    return DateTime.fromFormat(time, 'HH:mm').toLocaleString(
      DateTime.TIME_SIMPLE,
    )
  }

  const dt = DateTime.fromFormat(time, 'HH:mm', { zone })

  return formatClockTime(dt)
}

export const valueToRule = (
  zone: string,
  v: Value,
): OnCallNotificationRuleInput => ({
  time: v.time
    ? DateTime.fromISO(v.time).setZone(zone).toFormat('HH:mm')
    : undefined,
  weekdayFilter: v.time ? v.weekdayFilter : undefined,
  target: { type: 'slackChannel', id: v.slackChannelID || '' },
})

export const ruleToInput = (
  v: OnCallNotificationRule,
): OnCallNotificationRuleInput | null =>
  v
    ? {
        time: v.time,
        id: v.id,
        weekdayFilter: v.weekdayFilter,
        target: { type: v.target.type, id: v.target.id },
      }
    : null

export function mapErrors(
  mErr?: ApolloError | null,
  ...qErr: Array<ApolloError | undefined>
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

      if (e.field.startsWith('target')) {
        return { ...e, field: 'slackChannelID' }
      }

      dialogErrs.push(e)
      return null
    })
    .filter((e) => e !== null) as RuleFieldError[]

  return [dialogErrs, fieldErrs]
}

export function ruleSummary(zone: string, r?: OnCallNotificationRule): string {
  if (!r) return ''
  if (!r.time) return 'when on-call changes.'

  return `${weekdaySummary(r.weekdayFilter)} at ${formatTime(zone, r.time)}`
}
