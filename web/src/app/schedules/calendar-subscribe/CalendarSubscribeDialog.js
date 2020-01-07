import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { CheckCircleOutline as SuccessIcon } from '@material-ui/icons'
import FormDialog from '../../dialogs/FormDialog'
import CalenderSuccessForm from './CalendarSuccessForm'
import CalendarSubscribeForm from './CalendarSubscribeForm'

const MOCK_URL =
  'www.calendarlabs.com/ical-calendar/ics/22/Chicago_Cubs_-_MLB.ics'
const SUBTITLE =
  'Create a unique iCalendar subscription URL that can be used in your preferred calendar application.'

export default function CalendarSubscribeDialog(props) {
  const [complete, setComplete] = useState(false)
  const [value, setValue] = useState({
    name: '',
    schedule: props.scheduleID || null,
    reminderMinutes: [],
  })

  function submit() {
    setComplete(true)
  }

  const form = complete ? (
    <CalenderSuccessForm url={MOCK_URL} />
  ) : (
    <CalendarSubscribeForm
      disableSchedField={Boolean(props.scheduleID)}
      onChange={setValue}
      value={value}
    />
  )

  return (
    <FormDialog
      title={
        complete ? (
          <div
            style={{ color: 'green', display: 'flex', alignItems: 'center' }}
          >
            <SuccessIcon />
            &nbsp;Success!
          </div>
        ) : (
          'Create New Calendar Subscription'
        )
      }
      subTitle={SUBTITLE}
      onClose={props.onClose}
      alert={complete}
      primaryActionLabel={complete ? 'Done' : null}
      onSubmit={() => (complete ? props.onClose() : submit())}
      form={form}
    />
  )
}

CalendarSubscribeDialog.propTypes = {
  onClose: p.func.isRequired,
  scheduleID: p.string,
}
