import {
  gql,
  MutationResult,
  QueryResult,
  useMutation,
  useQuery,
} from '@apollo/client'
import { DateTime } from 'luxon'
import { OnCallNotificationRule } from '../../../schema'
import {
  Value,
  valueToRule,
  RuleFieldError,
  ruleSummary,
  NO_DAY,
  mapErrors,
  EVERY_DAY,
  formatLocalClockTime,
  ruleToInput,
} from './util'

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
export function useSubmit(
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

export type UpdateRuleState = {
  dialogErrors: Error[]
  fieldErrors: RuleFieldError[]

  busy: boolean
  submit: () => Promise<void>
}

export type UpsertRuleState = UpdateRuleState & {
  value: Value
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

export type DeleteRuleState = UpdateRuleState & {
  rule?: OnCallNotificationRule
  ruleSummary: string
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
