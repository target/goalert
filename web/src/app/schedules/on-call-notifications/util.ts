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

export type NotificationChannelType = 'SLACK' | 'WEBHOOK'

export type SlackFields = {
  slackChannelID: string | null
  slackUserGroup?: string
}

export type WebhookFields = {
  webhookURL: string | null
}

export type ChannelFields = SlackFields | WebhookFields

export function channelFieldsFromType(
  type: NotificationChannelType,
): ChannelFields {
  switch (type) {
    case 'SLACK':
      return {
        slackChannelID: null,
      }
    case 'WEBHOOK':
      return {
        webhookURL: null,
      }
    default:
      return {
        slackChannelID: null,
      }
  }
}

export function channelTypeFromTarget(
  target?: TargetInput,
): NotificationChannelType {
  if (
    !target ||
    target.type === 'slackChannel' ||
    target.type === 'slackUserGroup'
  ) {
    return 'SLACK'
  }
  if (target.type === 'chanWebhook') {
    return 'WEBHOOK'
  }
  return 'SLACK'
}

export function channelFieldsFromTarget(target?: TargetInput): ChannelFields {
  switch (target?.type) {
    case 'slackChannel':
      return {
        slackChannelID: target.id,
      }
    case 'slackUserGroup':
      return {
        slackChannelID: target.id.split(':')[1],
        slackUserGroup: target.id.split(':')[0],
      }
    case 'chanWebhook':
      return {
        webhookURL: target.id,
      }
    default:
      return {
        slackChannelID: null,
      }
  }
}

export type Value = {
  time: string | null
  weekdayFilter: WeekdayFilter
  type: NotificationChannelType
  channelFields: ChannelFields
}

export type RuleFieldError = {
  field: 'time' | 'weekdayFilter' | 'slackChannelID' | 'slackUserGroup' | 'type'
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

function targetFromChannelFields(fields?: ChannelFields): TargetInput {
  const target: TargetInput = {
    type: 'slackChannel',
    id: '',
  }
  if (!fields) {
    return target
  }
  if ('slackUserGroup' in fields) {
    target.type = 'slackUserGroup'
    target.id = `${fields.slackUserGroup}:${fields.slackChannelID}`
  } else if ('slackChannelID' in fields) {
    target.type = 'slackChannel'
    target.id = fields.slackChannelID ?? ''
  } else if ('webhookURL' in fields) {
    target.type = 'chanWebhook'
    target.id = fields.webhookURL ?? ''
  }
  return target
}

export const onCallValueToRuleInput = (
  zone: string,
  v: Value,
): OnCallNotificationRuleInput => ({
  time: v.time
    ? DateTime.fromISO(v.time).setZone(zone).toFormat('HH:mm')
    : undefined,
  weekdayFilter: v.time ? v.weekdayFilter : undefined,
  target: targetFromChannelFields(v.channelFields),
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

      if (e.field === 'targetTypeChanWebhook') {
        return { ...e, field: 'webhookURL' }
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
