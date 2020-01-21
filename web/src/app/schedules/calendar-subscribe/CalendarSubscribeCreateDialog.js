import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { useMutation } from '@apollo/react-hooks'
import gql from 'graphql-tag'
import FormDialog from '../../dialogs/FormDialog'
import CalendarSubscribeForm from './CalendarSubscribeForm'
import { getForm, FormTitle, getSubtitle } from './formHelper'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'

const mutation = gql`
  mutation($input: CreateUserCalendarSubscriptionInput!) {
    createUserCalendarSubscription(input: $input) {
      id
    }
  }
`

const MOCK_URL =
  'www.calendarlabs.com/ical-calendar/ics/22/Chicago_Cubs_-_MLB.ics'

const SUBTITLE =
  'Create a unique iCalendar subscription URL that can be used in your preferred calendar application.'

export default function CalendarSubscribeCreateDialog(props) {
  const [isComplete, setIsComplete] = useState(false)
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
        reminderMinutes: value.reminderMinutes.map(r => parseInt(r.value, 10)),
        disabled: false,
      },
    },
    onCompleted: () => setIsComplete(true),
  })

  const form = (
    <CalendarSubscribeForm
      disableSchedField={Boolean(props.scheduleID)}
      errors={fieldErrors(status.error)}
      loading={status.loading}
      onChange={setValue}
      value={value}
    />
  )

  return (
    <FormDialog
      title={FormTitle(isComplete, 'Create New Calendar Subscription')}
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
