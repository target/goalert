/* eslint-disable no-fallthrough */
import { ApolloError } from '@apollo/client'
import { DateTime } from 'luxon'

import {
  OnCallNotificationRule,
  OnCallNotificationRuleInput,
  TargetInput,
  WeekdayFilter,
} from '../../../schema'
import { allErrors, fieldErrors, nonFieldErrors } from '../../util/errutil'
import { weekdaySummary } from '../util'

export type NotificationChannelType = 'SLACK_CHANNEL' | 'SLACK_UG'

export type Value = {
  time: string | null
  weekdayFilter: WeekdayFilter
  type: NotificationChannelType
  channelField: string | null
  slackUserGroup?: string | null
}

// channelTypeFromTarget will return the NotificationChannelType based on the type
// of the target supplied. If the target is undefined or has an invalid type,
// SLACK_CHANNEL is returned by default.
export function channelTypeFromTarget(
  target?: TargetInput,
): NotificationChannelType {
  switch (target?.type) {
    case 'slackUserGroup':
      return 'SLACK_UG'
    case 'slackChannel':
    default:
      return 'SLACK_CHANNEL'
  }
}

// channelFieldsFromTarget will return an object with a channelField and possibly a
// slackUserGroup based on the target supplied. If the target is undefined or has
// an invalid type, an object with a null channelField is returned.
export function channelFieldsFromTarget(target?: TargetInput): {
  channelField: string | null
  slackUserGroup?: string | null
} {
  switch (target?.type) {
    case 'slackChannel':
      return {
        channelField: target.id,
      }
    case 'slackUserGroup':
      return {
        channelField: target.id.split(':')[1],
        slackUserGroup: target.id.split(':')[0],
      }
    default:
      return {
        channelField: null,
      }
  }
}

// targetFromValue will create a TargetInput object based on the value
// supplied.
function targetFromValue(value: Value): TargetInput {
  switch (value.type) {
    case 'SLACK_UG':
      return {
        type: 'slackUserGroup',
        id: `${value.slackUserGroup ?? ''}:${value.channelField ?? ''}`,
      }
    case 'SLACK_CHANNEL':
    default:
      return {
        type: 'slackChannel',
        id: value.channelField ?? '',
      }
  }
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
  target: targetFromValue(v),
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

export function onCallRuleSummary(
  zone: string,
  r?: OnCallNotificationRule,
): string {
  if (!r) return ''
  if (!r.time) return 'when on-call changes.'

  return `${weekdaySummary(r.weekdayFilter)} at ${formatZoneTime(zone, r.time)}`
}
