import React from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { Rule, withoutTypeName } from './util'
import { useURLParam } from '../../actions/hooks'
import { DateTime } from 'luxon'
import { weekdaySummary } from '../util'

function getDeleteSummary(r: Rule, scheduleZone: string): string {
  const prefix = `${r.target.name} will no longer be notified`
  if (!r.time) {
    return `${prefix} when on-call changes.`
  }

  const dt = DateTime.fromFormat(r.time, 'HH:mm', {
    zone: scheduleZone,
  })

  const timeStr = dt.toLocaleString(DateTime.TIME_SIMPLE)
  const localStr = dt.setZone('local').toLocaleString(DateTime.TIME_SIMPLE)

  const summary = `${prefix} ${weekdaySummary(r.weekdayFilter)} at ${timeStr}`
  if (timeStr === localStr) {
    return summary
  }

  return summary + ` (${localStr} ${dt.setZone('local').toFormat('ZZZZ')})`
}

const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
      onCallNotificationRules {
        id
        target {
          id
          type
          name
        }
        time
        weekdayFilter
      }
    }
  }
`
const mutation = gql`
  mutation ($input: SetScheduleOnCallNotificationRulesInput!) {
    setScheduleOnCallNotificationRules(input: $input)
  }
`

interface ScheduleOnCallNotificationsDeleteDialogProps {
  scheduleID: string
  ruleID: string
  onClose: () => void
}

export default function ScheduleOnCallNotificationsDeleteDialog(
  p: ScheduleOnCallNotificationsDeleteDialogProps,
): JSX.Element {
  const [displayZone] = useURLParam('tz', 'local')
  const { data, loading, error } = useQuery(query, {
    variables: { id: p.scheduleID },
  })

  const rules = (data?.schedule?.onCallNotificationRules ?? []).filter((r) => r)

  const rule = rules.find((r) => r.id === p.ruleID)
  const [mutate, mutationStatus] = useMutation(mutation, {
    variables: {
      input: {
        scheduleID: p.scheduleID,
        rules: rules.filter((r) => r.id !== p.ruleID).map(withoutTypeName),
      },
    },
    onCompleted: () => p.onClose(),
  })

  if (!loading && !rule) {
    return (
      <FormDialog
        alert
        title='No longer exists'
        onClose={() => p.onClose()}
        subTitle='That notification rule does not exist or is already deleted.'
      />
    )
  }

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={loading || mutationStatus.loading}
      errors={nonFieldErrors(error || mutationStatus.error)}
      subTitle={getDeleteSummary(rule, data?.schedule?.timeZone, displayZone)}
      onSubmit={() => mutate()}
      onClose={() => p.onClose()}
    />
  )
}
