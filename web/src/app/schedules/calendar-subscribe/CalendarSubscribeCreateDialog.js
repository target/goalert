import React, { useState } from 'react'
import { useMutation, gql } from '@apollo/client'
import { PropTypes as p } from 'prop-types'
import FormDialog from '../../dialogs/FormDialog'
import CalendarSubscribeForm from './CalendarSubscribeForm'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import { Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { CheckCircleOutline as SuccessIcon } from '@mui/icons-material'
import CalenderSuccessForm from './CalendarSuccessForm'

const mutation = gql`
  mutation ($input: CreateUserCalendarSubscriptionInput!) {
    createUserCalendarSubscription(input: $input) {
      id
      url
    }
  }
`

const useStyles = makeStyles((theme) => ({
  successIcon: {
    marginRight: theme.spacing(1),
  },
  successTitle: {
    display: 'flex',
    alignItems: 'center',
  },
}))

const SUBTITLE =
  'Create a unique iCalendar subscription URL that can be used in your preferred calendar application.'

export function getSubtitle(isComplete, defaultSubtitle) {
  const completedSubtitle =
    'Your subscription has been created! You can' +
    ' manage your subscriptions from your profile at any time.'

  return isComplete ? completedSubtitle : defaultSubtitle
}

export function getForm(isComplete, defaultForm, data) {
  return isComplete ? (
    <CalenderSuccessForm url={data.createUserCalendarSubscription.url} />
  ) : (
    defaultForm
  )
}

export default function CalendarSubscribeCreateDialog(props) {
  const classes = useStyles()

  const [value, setValue] = useState({
    name: '',
    scheduleID: props.scheduleID || null,
    reminderMinutes: [],
    fullSchedule: false,
  })

  const [createSubscription, status] = useMutation(mutation, {
    variables: {
      input: {
        scheduleID: value.scheduleID,
        name: value.name,
        reminderMinutes: [0], // default reminder at shift start time
        disabled: false,
        fullSchedule: value.fullSchedule,
      },
    },
  })

  const isComplete = Boolean(status?.data?.createUserCalendarSubscription?.url)

  const form = (
    <CalendarSubscribeForm
      errors={fieldErrors(status.error)}
      loading={status.loading}
      onChange={setValue}
      scheduleReadOnly={Boolean(props.scheduleID)}
      value={value}
    />
  )

  return (
    <FormDialog
      title={
        isComplete ? (
          <div className={classes.successTitle}>
            <SuccessIcon className={classes.successIcon} />
            <Typography>Success!</Typography>
          </div>
        ) : (
          'Create New Calendar Subscription'
        )
      }
      subTitle={getSubtitle(isComplete, SUBTITLE)}
      onClose={props.onClose}
      alert={isComplete}
      errors={nonFieldErrors(status.error)}
      primaryActionLabel={isComplete ? 'Done' : null}
      onSubmit={() => (isComplete ? props.onClose() : createSubscription())}
      form={getForm(isComplete, form, status.data)}
    />
  )
}

CalendarSubscribeCreateDialog.propTypes = {
  onClose: p.func.isRequired,
  scheduleID: p.string,
}
