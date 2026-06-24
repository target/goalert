import React, { ReactNode, useState } from 'react'
import { useMutation, gql } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import CalendarSubscribeForm, { CalSubFormValue } from './CalendarSubscribeForm'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import { Typography } from '@mui/material'
import { CheckCircleOutline as SuccessIcon } from '@mui/icons-material'
import CalenderSuccessForm from './CalendarSuccessForm'
import { UserCalendarSubscription } from '../../../schema'

const mutation = gql`
  mutation ($input: CreateUserCalendarSubscriptionInput!) {
    createUserCalendarSubscription(input: $input) {
      id
      url
    }
  }
`

const SUBTITLE =
  'Create a unique iCalendar subscription URL that can be used in your preferred calendar application.'

export function getSubtitle(
  isComplete: boolean,
  defaultSubtitle: string,
): string {
  const completedSubtitle =
    'Your subscription has been created! You can' +
    ' manage your subscriptions from your profile at any time.'

  return isComplete ? completedSubtitle : defaultSubtitle
}

export function getForm(
  isComplete: boolean,
  defaultForm: ReactNode,
  data: { createUserCalendarSubscription: UserCalendarSubscription },
): ReactNode {
  return isComplete ? (
    <CalenderSuccessForm url={data.createUserCalendarSubscription.url!} />
  ) : (
    defaultForm
  )
}

interface CalendarSubscribeCreateDialogProps {
  onClose: () => void
  scheduleID?: string
}

export default function CalendarSubscribeCreateDialog(
  props: CalendarSubscribeCreateDialogProps,
): ReactNode {
  const [value, setValue] = useState<CalSubFormValue>({
    name: '',
    scheduleID: props.scheduleID || null,
    reminderMinutes: [],
    fullSchedule: false,
  })

  const [status, commit] = useMutation(mutation)

  const isComplete = Boolean(status?.data?.createUserCalendarSubscription?.url)

  const form = (
    <CalendarSubscribeForm
      errors={fieldErrors(status.error)}
      loading={status.fetching}
      onChange={setValue}
      scheduleReadOnly={Boolean(props.scheduleID)}
      value={value}
    />
  )

  return (
    <FormDialog
      title={
        isComplete ? (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
            }}
          >
            <SuccessIcon sx={{ marginRight: (theme) => theme.spacing(1) }} />
            <Typography>Success!</Typography>
          </div>
        ) : (
          'Create New Calendar Subscription'
        )
      }
      subTitle={getSubtitle(isComplete, SUBTITLE)}
      onClose={props.onClose}
      alert={isComplete}
      loading={status.fetching}
      errors={nonFieldErrors(status.error)}
      primaryActionLabel={isComplete ? 'Done' : null}
      onSubmit={() =>
        isComplete
          ? props.onClose()
          : commit(
              {
                input: {
                  scheduleID: value.scheduleID,
                  name: value.name,
                  reminderMinutes: [0], // default reminder at shift start time
                  disabled: false,
                  fullSchedule: value.fullSchedule,
                },
              },
              { additionalTypenames: ['User'] },
            )
      }
      form={getForm(isComplete, form, status.data)}
    />
  )
}
