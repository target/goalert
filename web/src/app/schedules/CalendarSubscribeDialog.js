import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import {
  Button,
  FormHelperText,
  Grid,
  TextField,
  Typography,
} from '@material-ui/core'
import { CheckCircleOutline as SuccessIcon } from '@material-ui/icons'
import FormDialog from '../dialogs/FormDialog'
import { FormContainer, FormField } from '../forms'
import MaterialSelect from '../selection/MaterialSelect'
import { ScheduleSelect } from '../selection'

const MOCK_URL =
  'www.calendarlabs.com/ical-calendar/ics/22/Chicago_Cubs_-_MLB.ics'

export default function CalendarSubscribeDialog(props) {
  const [complete, setComplete] = useState(false)
  const [value, setValue] = useState({
    name: '',
    schedule: props.scheduleID || null,
    alarm: {
      label: '30 minutes before',
      value: '-P30M',
    },
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
    props.open && (
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
            'Create New Subscription'
          )
        }
        onClose={props.onClose}
        alert={complete}
        primaryActionLabel={complete ? 'Done' : null}
        onSubmit={() => (complete ? props.onClose() : submit())}
        form={form}
      />
    )
  )
}

CalendarSubscribeDialog.propTypes = {
  open: p.bool.isRequired,
  onClose: p.func.isRequired,
  scheduleID: p.string,
}

function CalendarSubscribeForm(props) {
  return (
    <FormContainer
      onChange={value => props.onChange(value)}
      optionalLabels
      value={props.value}
    >
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            name='name'
            label='Name'
            placeholder='My Outlook Calendar'
            required
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            component={ScheduleSelect}
            disabled={props.disableSchedField}
            fieldName='schedule'
            fullWidth
            required
            label='Schedule'
            name='schedule'
            InputLabelProps={{
              shrink: Boolean(props.value.schedule),
            }}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={MaterialSelect}
            name='alarm'
            label='Notify'
            required
            options={[
              {
                label: 'At time of shift',
                value: '-P0M',
              },
              {
                label: '5 minutes before',
                value: '-P5M',
              },
              {
                label: '10 minutes before',
                value: '-P10M',
              },
              {
                label: '30 minutes before',
                value: '-P30M',
              },
              {
                label: '1 hour before',
                value: '-P1H',
              },
              {
                label: '1 day before',
                value: '-P1D',
              },
            ]}
            InputLabelProps={{
              shrink: Boolean(props.value.alarm),
            }}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}

CalendarSubscribeForm.propTypes = {
  disableSchedField: p.bool,
  onChange: p.func.isRequired,
  value: p.object.isRequired,
}

export function CalenderSuccessForm(props) {
  const style = { display: 'flex', justifyContent: 'center' }
  return (
    <Grid container spacing={2}>
      <Grid item xs={12} style={style}>
        <Typography>
          Your subscription has been created! You can manage your subscriptions
          from your profile at anytime.
        </Typography>
      </Grid>
      <Grid item xs={12} style={style}>
        <Button
          color='primary'
          variant='contained'
          href={'webcal://' + MOCK_URL}
          style={{ marginLeft: '0.5em' }}
        >
          Subscribe
        </Button>
      </Grid>
      <Grid item xs={12}>
        <TextField
          value={props.url}
          onChange={() => {}}
          style={{ width: '100%' }}
        />
        <FormHelperText>
          Some applications require you copy and paste the URL directly
        </FormHelperText>
      </Grid>
    </Grid>
  )
}

CalenderSuccessForm.propTypes = {
  url: p.string.isRequired,
}
