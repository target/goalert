import React from 'react'
import { useQuery, useMutation } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { query, setMutation } from './ScheduleOnCallNotificationsList'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import { Rule, mapDataToInput, getDayNames } from './util'

function getDeleteSummary(r: Rule): string {
  const prefix = `${r.target.name} will no longer be notified`

  if (r.time && r.weekdayFilter) {
    return `${prefix} ${getDayNames(r.weekdayFilter)} at ${r.time}`
  }

  return `${prefix} when on-call changes.`
}

interface ScheduleOnCallNotificationDeleteDialogProps {
  rule: Rule
  scheduleID: string
  onClose: () => void
}

export default function ScheduleOnCallNotificationDeleteDialog(
  p: ScheduleOnCallNotificationDeleteDialogProps,
): JSX.Element {
  const { loading, error, data } = useQuery(query, {
    variables: {
      id: p.scheduleID,
    },
  })

  const [mutate, mutationStatus] = useMutation(setMutation, {
    variables: {
      input: {
        scheduleID: p.scheduleID,
        rules: mapDataToInput(
          data.schedule.onCallNotificationRules.filter(
            (nr: Rule) => nr.id !== p.rule.id,
          ),
        ),
      },
    },
    onCompleted: () => p.onClose(),
  })

  if (loading && !data?.schedule) return <Spinner />
  if (error) return <GenericError error={error.message} />

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={mutationStatus.loading}
      errors={nonFieldErrors(mutationStatus.error)}
      subTitle={getDeleteSummary(p.rule)}
      onSubmit={() => mutate()}
      onClose={() => p.onClose()}
    />
  )
}
