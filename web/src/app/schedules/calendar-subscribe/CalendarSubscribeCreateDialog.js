import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { useMutation } from '@apollo/react-hooks'
import gql from 'graphql-tag'
import FormDialog from '../../dialogs/FormDialog'
import CalendarSubscribeForm from './CalendarSubscribeForm'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import { makeStyles } from "@material-ui/core"
import { CheckCircleOutline as SuccessIcon } from "@material-ui/icons"
import CalenderSuccessForm from "./CalendarSuccessForm"

const mutation = gql`
  mutation($input: CreateUserCalendarSubscriptionInput!) {
    createUserCalendarSubscription(input: $input) {
      id
    }
  }
`

const useStyles = makeStyles(theme => ({
  successIcon: {
    marginRight: theme.spacing(1),
  },
  successTitle: {
    color: 'green',
    display: 'flex',
    alignItems: 'center',
  },
}))

const MOCK_URL =
  'www.calendarlabs.com/ical-calendar/ics/22/Chicago_Cubs_-_MLB.ics'

const SUBTITLE =
  'Create a unique iCalendar subscription URL that can be used in your preferred calendar application.'

export function getSubtitle(isComplete, defaultSubtitle) {
  const completedSubtitle =
    'Your subscription has been created! You can' +
    ' manage your subscriptions from your profile at any time.'

  return isComplete ? completedSubtitle : defaultSubtitle
}

export function getForm(isComplete, defaultForm, url) {
  return isComplete ? <CalenderSuccessForm url={url} /> : defaultForm
}

export default function CalendarSubscribeCreateDialog(props) {
  const classes = useStyles()

  const [value, setValue] = useState({
    name: '',
    scheduleID: props.scheduleID || null,
    reminderMinutes: [],
  })

  const [createSubscription, status] = useMutation(mutation, {
    variables: {
      input: {
        scheduleID: value.scheduleID,
        name: value.name,
        reminderMinutes: value.reminderMinutes.map(r => r && r.value),
        disabled: false,
      },
    },
  })

  // todo: status?.data?.href
  const isComplete = status.called && status.data && !status.error && !status.loading

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
      title={isComplete ? (
        <div className={classes.successTitle}>
          <SuccessIcon className={classes.successIcon} />
          Success!
        </div>
      ) : 'Create New Calendar Subscription'}
      subTitle={getSubtitle(isComplete, SUBTITLE)}
      onClose={props.onClose}
      alert={isComplete}
      errors={nonFieldErrors(status.error)}
      primaryActionLabel={isComplete ? 'Done' : null}
      onSubmit={() => (isComplete ? props.onClose() : createSubscription())}
      form={getForm(isComplete, form, MOCK_URL)}
    />
  )
}

CalendarSubscribeCreateDialog.propTypes = {
  onClose: p.func.isRequired,
  scheduleID: p.string,
}
