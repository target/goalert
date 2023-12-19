import {
  Checkbox,
  FormControlLabel,
  Grid,
  Radio,
  RadioGroup,
  TextField,
  Typography,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { DateTime } from 'luxon'
import React from 'react'

import {
  DestinationInput,
  OnCallNotificationRuleInput,
  TargetType,
} from '../../../schema'
import { FormContainer, FormField } from '../../forms'
import { renderMenuItem } from '../../selection/DisableableMenuItem'
import { ISOTimePicker } from '../../util/ISOPickers'
import { Time } from '../../util/Time'
import { useScheduleTZ } from '../useScheduleTZ'
import { EVERY_DAY, NO_DAY, RuleFieldError } from './util'
import { useSchedOnCallNotifyTypes } from '../../util/useDestinationTypes'
import DestinationField from '../../selection/DestinationField'

const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

export type FormValue = Omit<OnCallNotificationRuleInput, 'target'> & {
  dest: DestinationInput
}

interface ScheduleOnCallNotificationsFormProps {
  scheduleID: string
  value: FormValue
  errors: RuleFieldError[]
  onChange: (val: FormValue) => void
}

const useStyles = makeStyles({
  margin0: { margin: 0 },
  tzNote: { fontStyle: 'italic' },
})

export default function ScheduleOnCallNotificationsForm(
  props: ScheduleOnCallNotificationsFormProps,
): JSX.Element {
  const { scheduleID, ...formProps } = props
  const classes = useStyles()

  const destinationTypes = useSchedOnCallNotifyTypes()

  const { zone } = useScheduleTZ(scheduleID)

  const handleRuleChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
    if (e.target.value === 'on-change') {
      props.onChange({ ...formProps.value, time: null, weekdayFilter: NO_DAY })
      return
    }
    props.onChange({
      ...props.value,
      weekdayFilter: EVERY_DAY,
      time: DateTime.fromObject({ hour: 9 }, { zone }).toISO(),
    })
  }

  const handleTypeChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
    const newType = e.target.value as TargetType
    if (props.value.dest.type === newType) return

    props.onChange({
      ...props.value,
      dest: {
        type: newType,
        values: [],
      },
    })
  }

  return (
    <FormContainer {...formProps}>
      <Grid container spacing={2} direction='column'>
        <Grid item>
          <RadioGroup
            name='ruleType'
            value={formProps.value.time ? 'time-of-day' : 'on-change'}
            onChange={handleRuleChange}
          >
            <FormControlLabel
              data-cy='notify-on-change'
              label='Notify when on-call changes'
              value='on-change'
              control={<Radio />}
            />
            <FormControlLabel
              data-cy='notify-at-time'
              label='Notify at a specific day and time every week'
              value='time-of-day'
              control={<Radio />}
            />
          </RadioGroup>
        </Grid>
        {props.value.time && (
          <Grid item xs={12}>
            <Typography color='textSecondary' className={classes.tzNote}>
              Times shown in schedule timezone ({zone})
            </Typography>
          </Grid>
        )}
        <Grid item container spacing={2} alignItems='center'>
          <Grid item xs={12} sm={5} md={4}>
            <FormField
              component={ISOTimePicker}
              timeZone={zone}
              fullWidth
              name='time'
              disabled={!props.value.time}
              required={!!props.value.time}
              hint={
                <Time
                  format='clock'
                  time={props.value.time}
                  suffix=' in local time'
                />
              }
            />
          </Grid>
          <Grid item xs={12} sm={7} md={8}>
            <Grid container justifyContent='space-between'>
              {days.map((day, i) => (
                <FormControlLabel
                  key={i}
                  label={day}
                  labelPlacement='top'
                  classes={{ labelPlacementTop: classes.margin0 }}
                  control={
                    <FormField
                      noError
                      component={Checkbox}
                      checkbox
                      name={`weekdayFilter[${i}]`}
                      disabled={!props.value.time}
                    />
                  }
                />
              ))}
            </Grid>
          </Grid>
        </Grid>
        <Grid item>
          <TextField
            fullWidth
            value={props.value.dest.type}
            required
            name='notificationType'
            label='Type'
            select
            onChange={handleTypeChange}
          >
            {destinationTypes.map((t) =>
              renderMenuItem({
                value: t.type,
                label: t.name,
                disabled: !t.enabled,
                disabledMessage: t.enabled ? '' : t.disabledMessage,
              }),
            )}
          </TextField>
        </Grid>
        <Grid item>
          <FormField
            component={DestinationField}
            fullWidth
            name='dest.values'
            destType={formProps.value.dest.type}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
