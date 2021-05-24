import React from 'react'
import { ApolloError, useQuery, useMutation } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { query, setMutation } from './ScheduleOnCallNotifications'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'

interface ScheduleOnCallNotificationDeleteDialogProps {
  id: string
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

  let name = ''
  const rulesAfterDelete = data.schedule.notificationRules.filter((nr) => {
    if (nr.id === p.id) {
      name = nr.channel
      return true
    }
    return false
  })

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={mutationStatus.loading}
      errors={nonFieldErrors(mutationStatus.error as ApolloError)}
      subTitle={name + ' will no longer be notified of on-call updates.'}
      onSubmit={() =>
        mutate({
          variables: {
            input: {
              scheduleID: p.scheduleID,
              rules: rulesAfterDelete,
            },
          },
        })
      }
      onClose={() => p.onClose()}
    />
  )
}
