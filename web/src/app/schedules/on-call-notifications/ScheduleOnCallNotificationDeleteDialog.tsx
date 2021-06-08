import React from 'react'
import { ApolloError, useQuery, useMutation } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { query, setMutation } from './ScheduleOnCallNotificationsList'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import { Rule, mapDataToInput, getDayNames } from './util'

function getDeleteSummary(r: Rule): string {
  const prefix = `${r.target.name} will no longer be notified`

  if (r.time && r.weekdayFilter) {
    const everyday = [true, true, true, true, true, true, true]
    const isEverday = r.weekdayFilter.every((val, i) => val === everyday[i])
    if (isEverday) return prefix + ' everyday at ' + r.time

    const weekdays = [false, true, true, true, true, true, false]
    const isWeekdays = r.weekdayFilter.every((val, i) => val === weekdays[i])
    if (isWeekdays) return prefix + ' on weekdays at ' + r.time

    return prefix + ' on ' + getDayNames(r.weekdayFilter) + ' at ' + r.time
  }

  return `${prefix} will no longer be notified when on-call changes.`
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
      errors={nonFieldErrors(mutationStatus.error as ApolloError)}
      subTitle={getDeleteSummary(p.rule)}
      onSubmit={() => mutate()}
      onClose={() => p.onClose()}
    />
  )
}
