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
import React, { useEffect } from 'react'

import { FormContainer, FormField } from '../../forms'
import { renderMenuItem } from '../../selection/DisableableMenuItem'
import { ISOTimePicker } from '../../util/ISOPickers'
import { Time } from '../../util/Time'
import { useScheduleTZ } from '../useScheduleTZ'
import { EVERY_DAY, NO_DAY } from './util'
import { useSchedOnCallNotifyTypes } from '../../util/RequireConfig'
import { DestinationInput, WeekdayFilter } from '../../../schema'
import DestinationField from '../../selection/DestinationField'

const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

export type Value = {
  time: string | null
  weekdayFilter: WeekdayFilter
  dest: DestinationInput
}

interface ScheduleOnCallNotificationsFormProps {
  scheduleID: string
  value: Value
  onChange: (val: Value) => void
  disablePortal?: boolean

  destTypeError?: string
  destFieldErrors?: Readonly<Record<string, string>>
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

export default function ScheduleOnCallNotificationsForm(
  props: ScheduleOnCallNotificationsFormProps,
): JSX.Element {
  const { scheduleID, ...formProps } = props
  const classes = useStyles()
  const { zone } = useScheduleTZ(scheduleID)
  const destinationTypes = useSchedOnCallNotifyTypes()
  const currentType = destinationTypes.find(
    (d) => d.type === props.value.dest.type,
  )

  const [ruleType, setRuleType] = React.useState<'on-change' | 'time-of-day'>(
    props.value.time ? 'time-of-day' : 'on-change',
  )
  const [lastTime, setLastTime] = React.useState<string | null>(
    props.value.time,
  )
  const [lastFilter, setLastFilter] = React.useState<WeekdayFilter>(
    props.value.weekdayFilter,
  )
  useEffect(() => {
    if (!props.value.time) return
    setLastTime(props.value.time)
    setLastFilter(props.value.weekdayFilter)
  }, [props.value.time, props.value.weekdayFilter])

  if (!currentType) throw new Error('invalid destination type')

  const handleRuleChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
    if (e.target.value === 'on-change') {
      setRuleType('on-change')
      props.onChange({ ...formProps.value, time: null, weekdayFilter: NO_DAY })
      return
    }

    setRuleType('time-of-day')
    props.onChange({
      ...props.value,
      weekdayFilter: lastTime ? lastFilter : EVERY_DAY,
      time: lastTime || DateTime.fromObject({ hour: 9 }, { zone }).toISO(),
    })
  }

  return (
    <FormContainer {...formProps} optionalLabels>
      <Grid container spacing={2} direction='column'>
        <Grid item>
          <RadioGroup
            name='ruleType'
            value={ruleType}
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
          <TextField
            fullWidth
            name='dest.type'
            label='Destination Type'
            select
            SelectProps={{ MenuProps: { disablePortal: props.disablePortal } }}
            value={props.value.dest.type}
            onChange={(e) =>
              props.onChange({
                ...props.value,
                dest: { type: e.target.value, args: {} },
              })
            }
            error={!!props.destTypeError}
            helperText={props.destTypeError}
          >
            {destinationTypes.map((t) =>
              renderMenuItem({
                label: t.name,
                value: t.type,
                disabled: !t.enabled,
                disabledMessage: t.enabled ? '' : 'Disabled by administrator.',
              }),
            )}
          </TextField>
        </Grid>
        <Grid item xs={12}>
          <DestinationField
            destType={props.value.dest.type}
            fieldErrors={props.destFieldErrors}
            value={props.value.dest.args || {}}
            onChange={(newValue) =>
              props.onChange({
                ...props.value,
                dest: { ...props.value.dest, args: newValue },
              })
            }
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
