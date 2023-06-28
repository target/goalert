import {
  gql,
  MutationResult,
  QueryResult,
  useMutation,
  useQuery,
} from '@apollo/client'
import { DateTime } from 'luxon'
import { OnCallNotificationRule, Schedule } from '../../../schema'
import {
  EVERY_DAY,
  formatLocalClockTime,
  mapOnCallErrors,
  NO_DAY,
  onCallRuleSummary,
  onCallRuleToInput,
  onCallValueToRuleInput,
  RuleFieldError,
  Value,
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
  q: QueryResult<Schedule, { id: string }>
  zone: string
  rules: OnCallNotificationRule[]
}

export function useOnCallRulesData(scheduleID: string): _Data {
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
export function useSetOnCallRulesSubmit(
  scheduleID: string,
  zone: string,
  ...rules: Array<OnCallNotificationRule | Value | null>
): _Submit {
  const variables = {
    input: {
      scheduleID,

      // map value and rules into input format
      rules: rules
        .map((r) => {
          if (r === null) return null
          if ('targetID' in r) {
            return onCallValueToRuleInput(zone, r)
          }

          return onCallRuleToInput(r as OnCallNotificationRule)
        })

        // remove any null values
        .filter((r) => r),
    },
  }

  const [submit, m] = useMutation(updateMutation, {
    variables,
  })
  return { m, submit: () => submit().then(() => {}) }
}

export type UpdateOnCallRuleState = {
  dialogErrors: Error[]
  fieldErrors: RuleFieldError[]

  busy: boolean
  submit: () => Promise<void>
}

export type UpsertOnCallRuleState = UpdateOnCallRuleState & {
  value: Value
}

export function useEditOnCallRule(
  scheduleID: string,
  ruleID: string,
  value: Value | null,
): UpsertOnCallRuleState {
  const { q, zone, rules } = useOnCallRulesData(scheduleID)

  const rule = rules.find((r) => r.id === ruleID)
  const newValue: Value = value || {
    time: rule?.time
      ? DateTime.fromFormat(rule.time, 'HH:mm', { zone }).toISO()
      : null,
    weekdayFilter: rule?.time ? rule.weekdayFilter || EVERY_DAY : NO_DAY,
    type: rule?.target?.type ?? 'slackChannel',
    targetID: rule?.target?.id ?? null,
  }
  const { m, submit } = useSetOnCallRulesSubmit(
    scheduleID,
    zone,
    newValue,
    ...rules.filter((r) => r.id !== ruleID),
  )

  const [dialogErrors, fieldErrors] = mapOnCallErrors(m.error, q.error)
  return {
    dialogErrors,
    fieldErrors,
    busy: (q.loading && !zone) || m.loading,
    value: newValue,
    submit,
  }
}

export type DeleteOnCallRuleState = UpdateOnCallRuleState & {
  rule?: OnCallNotificationRule
  ruleSummary: string
}

export function useDeleteOnCallRule(
  scheduleID: string,
  ruleID: string,
): DeleteOnCallRuleState {
  const { q, zone, rules } = useOnCallRulesData(scheduleID)
  const { m, submit } = useSetOnCallRulesSubmit(
    scheduleID,
    zone,
    ...rules.filter((r) => r.id !== ruleID),
  )
  const rule = rules.find((r) => r.id === ruleID)

  // treat all field errors as dialog errors for delete
  const [dialogErrors, fieldErrors] = mapOnCallErrors(null, m.error, q.error)
  return {
    dialogErrors,
    fieldErrors,
    busy: (q.loading && !zone) || m.loading,
    rule,
    ruleSummary: onCallRuleSummary(zone, rule),
    submit,
  }
}
