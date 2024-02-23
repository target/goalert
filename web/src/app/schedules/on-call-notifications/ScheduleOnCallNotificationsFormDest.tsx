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

import { FormContainer, FormField } from '../../forms'
import { renderMenuItem } from '../../selection/DisableableMenuItem'
import { ISOTimePicker } from '../../util/ISOPickers'
import { Time } from '../../util/Time'
import { useScheduleTZ } from '../useScheduleTZ'
import { EVERY_DAY, NO_DAY } from './util'
import { useSchedOnCallNotifyTypes } from '../../util/RequireConfig'
import { DestinationInput, WeekdayFilter } from '../../../schema'
import DestinationField from '../../selection/DestinationField'
import {
  DestFieldValueError,
  KnownError,
  isDestFieldError,
  isInputFieldError,
} from '../../util/errtypes'

const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

export type Value = {
  time: string | null
  weekdayFilter: WeekdayFilter
  dest: DestinationInput
}

interface ScheduleOnCallNotificationsFormProps {
  scheduleID: string
  value: Value
  errors?: Array<KnownError | DestFieldValueError>
  onChange: (val: Value) => void
  disablePortal?: boolean
}

const useStyles = makeStyles({
  margin0: { margin: 0 },
  tzNote: { fontStyle: 'italic' },
})

export const errorPaths = (prefix = '*'): string[] => [
  `${prefix}.time`,
  `${prefix}.weedkayFilter`,
  `${prefix}.dest.type`,
  `${prefix}.dest`,
]

export default function ScheduleOnCallNotificationsFormDest(
  props: ScheduleOnCallNotificationsFormProps,
): JSX.Element {
  const { scheduleID, ...formProps } = props
  const classes = useStyles()
  const { zone } = useScheduleTZ(scheduleID)
  const destinationTypes = useSchedOnCallNotifyTypes()
  const currentType = destinationTypes.find(
    (d) => d.type === props.value.dest.type,
  )

  if (!currentType) throw new Error('invalid destination type')

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

  return (
    <FormContainer
      {...formProps}
      errors={props.errors?.filter(isInputFieldError).map((e) => {
        let field = e.path[e.path.length - 1].toString()
        if (field === 'type') field = 'dest.type'
        return {
          // need to convert to FormContainer's error format
          message: e.message,
          field,
        }
      })}
      optionalLabels
    >
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
        <Grid item xs={12} sm={12} md={6}>
          <FormField
            fullWidth
            name='dest.type'
            label='Destination Type'
            required
            select
            disablePortal={props.disablePortal}
            component={TextField}
          >
            {destinationTypes.map((t) =>
              renderMenuItem({
                label: t.name,
                value: t.type,
                disabled: !t.enabled,
                disabledMessage: t.enabled ? '' : t.disabledMessage,
              }),
            )}
          </FormField>
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            name='value'
            fieldName='dest.values'
            required
            destType={props.value.dest.type}
            component={DestinationField}
            destFieldErrors={props.errors?.filter(isDestFieldError)}
          />
        </Grid>
        <Grid item xs={12}>
          <Typography variant='caption'>
            {currentType?.userDisclaimer}
          </Typography>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
