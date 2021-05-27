import React from 'react'
import { ApolloError, useQuery, useMutation } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { query, setMutation } from './ScheduleOnCallNotificationsList'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import { Rule } from './ScheduleOnCallNotificationAction'

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
    nextFetchPolicy: 'cache-first',
  })
  const [mutate, mutationStatus] = useMutation(setMutation)

  if (loading && !data?.schedule) return <Spinner />
  if (error) return <GenericError error={error.message} />

  function handleOnSubmit(): void {
    mutate({
      variables: {
        input: {
          scheduleID: p.scheduleID,
          rules: data.schedule.onCallNotificationRules.filter(
            (nr: Rule) => nr.id !== p.rule.id,
          ),
        },
      },
    })
  }

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={mutationStatus.loading}
      errors={nonFieldErrors(mutationStatus.error as ApolloError)}
      subTitle={
        p.rule.target.name + ' will no longer be notified of on-call updates.'
      }
      onSubmit={() => handleOnSubmit()}
      onClose={() => p.onClose()}
    />
  )
}
