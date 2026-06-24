import React, { ReactNode, useState } from 'react'
import { useQuery, useMutation, gql } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import CalendarSubscribeForm, { CalSubFormValue } from './CalendarSubscribeForm'
import { GenericError, ObjectNotFound } from '../../error-pages'
import _ from 'lodash'
import Spinner from '../../loading/components/Spinner'
import { fieldErrors } from '../../util/errutil'
import { UserCalendarSubscription } from '../../../schema'

const query = gql`
  query ($id: ID!) {
    userCalendarSubscription(id: $id) {
      id
      name
      scheduleID
      fullSchedule
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateUserCalendarSubscriptionInput!) {
    updateUserCalendarSubscription(input: $input)
  }
`

interface CalendarSubscribeEditDialogContentProps {
  data: UserCalendarSubscription
  onClose: () => void
}

export function CalendarSubscribeEditDialogContent(
  props: CalendarSubscribeEditDialogContentProps,
): ReactNode {
  const { data, onClose } = props

  // set default values from retrieved data
  const [value, setValue] = useState<CalSubFormValue>({
    name: _.get(data, 'name', ''),
    scheduleID: _.get(data, 'scheduleID', null),
    fullSchedule: _.get(data, 'fullSchedule', false),
    reminderMinutes: _.get(data, 'reminderMinutes', []),
  })

  // setup the mutation
  const [updateSubscriptionStatus, commit] = useMutation(mutation)

  return (
    <FormDialog
      title='Edit Calendar Subscription'
      onClose={onClose}
      loading={updateSubscriptionStatus.fetching}
      onSubmit={() =>
        commit(
          {
            input: {
              id: props.data.id,
              name: value.name,
              fullSchedule: value.fullSchedule,
            },
          },
          { additionalTypenames: ['UserCalendarSubscription'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }
      form={
        <CalendarSubscribeForm
          errors={fieldErrors(updateSubscriptionStatus.error)}
          loading={updateSubscriptionStatus.fetching}
          onChange={setValue}
          value={value}
          scheduleReadOnly
        />
      }
    />
  )
}

interface CalendarSubscribeEditDialogProps {
  calSubscriptionID: string
  onClose: () => void
}

/*
 * Load edit data here before rendering edit content to
 * avoid breaking any rules of hooks
 */
export default function CalendarSubscribeEditDialog(
  props: CalendarSubscribeEditDialogProps,
): ReactNode {
  const [{ data, fetching, error }] = useQuery({
    query,
    variables: { id: props.calSubscriptionID },
  })

  if (error) return <GenericError error={error.message} />
  if (!_.get(data, 'userCalendarSubscription.id')) {
    return fetching ? <Spinner /> : <ObjectNotFound />
  }

  return (
    <CalendarSubscribeEditDialogContent
      data={data.userCalendarSubscription}
      onClose={props.onClose}
    />
  )
}
