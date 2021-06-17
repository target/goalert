import {
  gql,
  useQuery,
  useMutation,
  QueryResult,
  MutationResult,
  ApolloError,
} from '@apollo/client'
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

export type UpdateRuleState = {
  dialogErrors: Error[]
  fieldErrors: RuleFieldError[]

  busy: boolean
  submit: () => Promise<void>
}

export type UpsertRuleState = UpdateRuleState & {
  value: Value
}

export type DeleteRuleState = UpdateRuleState & {
  rule?: OnCallNotificationRule
  ruleSummary: string
}

export const EVERY_DAY = new Array(7).fill(true) as WeekdayFilter
export const NO_DAY = new Array(7).fill(false) as WeekdayFilter

const schedTZQuery = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
    }
  }
`

const rulesQuery = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
      rules: onCallNotificationRules {
        id
        target {
          id
          name
          type
        }
        time
        weekdayFilter
      }
    }
  }
`

const updateMutation = gql`
  mutation ($input: SetScheduleOnCallNotificationRulesInput!) {
    setScheduleOnCallNotificationRules(input: $input)
  }
`

function formatLocalClockTime(dt: DateTime): string {
  const zoneStr = dt.toLocaleString(DateTime.TIME_SIMPLE)
  const localStr = dt.setZone('local').toLocaleString(DateTime.TIME_SIMPLE)

  if (zoneStr === localStr) {
    return ''
  }

  return `${localStr} ${dt.setZone('local').toFormat('ZZZZ')}`
}

function formatClockTime(dt: DateTime): string {
  const zoneStr = dt.toLocaleString(DateTime.TIME_SIMPLE)
  const localStr = formatLocalClockTime(dt)
  if (!localStr) {
    return zoneStr
  }

  return `${zoneStr} (${localStr})`
}

export function useFormatScheduleLocalISOTime(
  scheduleID: string,
): [(isoTime: string | null) => string, string] {
  const { data } = useQuery(schedTZQuery, {
    variables: { id: scheduleID },
  })
  const tz = data?.schedule?.timeZone
  return [
    (isoTime: string | null) => {
      if (!tz || !isoTime) return ''
      return formatLocalClockTime(DateTime.fromISO(isoTime).setZone(tz))
    },
    tz,
  ]
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

type _Data = {
  q: QueryResult
  zone: string
  rules: OnCallNotificationRule[]
}

export function useRulesData(scheduleID: string): _Data {
  const q = useQuery(rulesQuery, { variables: { id: scheduleID } })
  const zone = q.data?.schedule?.timeZone || ''
  const rules: OnCallNotificationRule[] = (
    q.data?.schedule?.rules || []
  ).filter((r?: OnCallNotificationRule) => r)
  return { q, zone, rules }
}
type _Submit = {
  m: MutationResult
  submit: () => Promise<void>
}

const valueToRule = (zone: string, v: Value): OnCallNotificationRuleInput => ({
  time: v.time
    ? DateTime.fromISO(v.time).setZone(zone).toFormat('HH:mm')
    : undefined,
  weekdayFilter: v.time ? v.weekdayFilter : undefined,
  target: { type: 'slackChannel', id: v.slackChannelID || '' },
})

const ruleToInput = (
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

function useSubmit(
  scheduleID: string,
  zone: string,
  ...rules: Array<OnCallNotificationRule | Value | null>
): _Submit {
  const [submit, m] = useMutation(updateMutation, {
    variables: {
      input: {
        scheduleID,

        // map value and rules into input format
        rules: rules
          .map((r) => {
            if (r === null) return null
            if ('slackChannelID' in r) {
              return valueToRule(zone, r)
            }

            return ruleToInput(r as OnCallNotificationRule)
          })

          // remove any null values
          .filter((r) => r),
      },
    },
  })
  return { m, submit: () => submit().then(() => {}) }
}

function mapErrors(
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

export function useCreateRule(
  scheduleID: string,
  value: Value | null,
): UpsertRuleState {
  const { q, zone, rules } = useRulesData(scheduleID)

  const newValue: Value = value || {
    time: null,
    weekdayFilter: NO_DAY,
    slackChannelID: null,
  }
  const { m, submit } = useSubmit(scheduleID, zone, newValue, ...rules)

  const [dialogErrors, fieldErrors] = mapErrors(m.error, q.error)
  return {
    dialogErrors,
    fieldErrors,
    busy: (q.loading && !zone) || m.loading,
    value: newValue,
    submit,
  }
}

export function useEditRule(
  scheduleID: string,
  ruleID: string,
  value: Value | null,
): UpsertRuleState {
  const { q, zone, rules } = useRulesData(scheduleID)

  const rule = rules.find((r) => r.id === ruleID)
  const newValue: Value = value || {
    time: rule?.time
      ? DateTime.fromFormat(rule.time, 'HH:mm', { zone }).toISO()
      : null,
    weekdayFilter: rule?.time ? rule.weekdayFilter || EVERY_DAY : NO_DAY,
    slackChannelID: rule?.target.id || null,
  }
  const { m, submit } = useSubmit(
    scheduleID,
    zone,
    newValue,
    ...rules.filter((r) => r.id !== ruleID),
  )

  const [dialogErrors, fieldErrors] = mapErrors(m.error, q.error)
  return {
    dialogErrors,
    fieldErrors,
    busy: (q.loading && !zone) || m.loading,
    value: newValue,
    submit,
  }
}

export function ruleSummary(zone: string, r?: OnCallNotificationRule): string {
  if (!r) return ''
  if (!r.time) return 'when on-call changes.'

  return `${weekdaySummary(r.weekdayFilter).toLowerCase()} at ${formatTime(
    zone,
    r.time,
  )}`
}

export function useDeleteRule(
  scheduleID: string,
  ruleID: string,
): DeleteRuleState {
  const { q, zone, rules } = useRulesData(scheduleID)
  const { m, submit } = useSubmit(
    scheduleID,
    zone,
    ...rules.filter((r) => r.id !== ruleID),
  )
  const rule = rules.find((r) => r.id === ruleID)

  // treat all field errors as dialog errors for delete
  const [dialogErrors, fieldErrors] = mapErrors(null, m.error, q.error)
  return {
    dialogErrors,
    fieldErrors,
    busy: (q.loading && !zone) || m.loading,
    rule,
    ruleSummary: ruleSummary(zone, rule),
    submit,
  }
}
