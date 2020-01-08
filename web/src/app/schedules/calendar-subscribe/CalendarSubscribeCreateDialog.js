import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import FormDialog from '../../dialogs/FormDialog'
import CalendarSubscribeForm from './CalendarSubscribeForm'
import { getForm, FormTitle, getSubtitle } from './formHelper'

const MOCK_URL =
  'www.calendarlabs.com/ical-calendar/ics/22/Chicago_Cubs_-_MLB.ics'

const SUBTITLE =
  'Create a unique iCalendar subscription URL that can be used in your preferred calendar application.'

export default function CalendarSubscribeCreateDialog(props) {
  const [isComplete, setIsComplete] = useState(false)
  const [value, setValue] = useState({
    name: '',
    schedule: props.scheduleID || null,
    reminderMinutes: [],
  })

  function submit() {
    setIsComplete(true)
  }

  const form = (
    <CalendarSubscribeForm
      disableSchedField={Boolean(props.scheduleID)}
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
      primaryActionLabel={isComplete ? 'Done' : null}
      onSubmit={() => (isComplete ? props.onClose() : submit())}
      form={getForm(isComplete, form, MOCK_URL)}
    />
  )
}

CalendarSubscribeCreateDialog.propTypes = {
  onClose: p.func.isRequired,
  scheduleID: p.string,
}
