import React from 'react'
import { FormContainer, FormField } from '../forms'
import {
  Grid,
  Checkbox,
  Table,
  TableHead,
  TableRow,
  TableCell,
  Hidden,
  IconButton,
  TableBody,
  MenuItem,
  TextField,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { UserSelect, RotationSelect } from '../selection'
import { startCase } from 'lodash'
import { Add, Trash } from '../icons'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import { ISOTimePicker } from '../util/ISOPickers'
import { DateTime } from 'luxon'
import { useURLParam } from '../actions'
import { ScheduleTargetInput, TargetType } from '../../schema'
import { FieldError } from '../util/errutil'
import { ElementType } from '../../global'

const days = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
] as const
type DayName = ElementType<typeof days>

// extends FormContainerProps
interface ScheduleRuleFormProps {
  // FormContainerProps
  disabled?: boolean
  errors: FieldError[]
  onChange: (value: ScheduleTargetInput) => void
  value: ScheduleTargetInput

  // ScheduleRuleFormProps
  targetType: TargetType
  targetDisabled?: boolean
  scheduleID: string
}

// renderDaysValue abbreviates an array of day names
// e.g. ["Monday", "Friday"] => "Mon, Fri"
// e.g. ["Monday", "Tuesday", "Friday"] => "Mon-Tues, Fri"
// TODO unit tests
const renderDaysValue = (value: DayName[]): string => {
  const parts: string[] = []
  let start = ''
  let last = ''
  let lastIdx = -1

  function flush(): void {
    if (lastIdx === -1) return
    if (start === last) {
      parts.push(start)
    } else {
      parts.push(start + '—' + last)
    }
    lastIdx = -1
  }

  days.forEach((day, idx) => {
    const enabled = value.includes(day)
    if (lastIdx === -1 && enabled) {
      start = day.substr(0, 3)
      last = start
      lastIdx = idx
    } else if (enabled) {
      lastIdx = idx
      last = day.substr(0, 3)
    } else if (lastIdx !== -1) {
      flush()
    }
  })

  flush()
  return parts.join(',')
}

const useStyles = makeStyles(() => {
  return {
    noPadding: {
      padding: 0,
    },
    dayFilter: {
      padding: 0,
      paddingRight: '1em',
    },
    startEnd: {
      padding: 0,
      minWidth: '6em',
      paddingRight: '1em',
    },
    tzNote: {
      display: 'flex',
      alignItems: 'center',
    },
    italic: {
      fontStyle: 'italic',
    },
  }
})

function ScheduleRuleForm(props: ScheduleRuleFormProps): JSX.Element {
  const { targetDisabled = false, targetType, scheduleID, ...formProps } = props
  const classes = useStyles()
  const [displayTZ] = useURLParam('tz', 'local')

  function renderRuleField(idx: number): JSX.Element {
    return (
      <TableRow key={idx}>
        <TableCell className={classes.startEnd}>
          <FormField
            fullWidth
            noError
            component={ISOTimePicker}
            required
            label=''
            name={`rules[${idx}].start`}
          />
        </TableCell>
        <TableCell className={classes.startEnd}>
          <FormField
            fullWidth
            noError
            component={ISOTimePicker}
            required
            label=''
            name={`rules[${idx}].end`}
          />
        </TableCell>
        <Hidden smDown>
          {days.map((_, dayIdx) => (
            <TableCell key={dayIdx} padding='checkbox'>
              <FormField
                noError
                className={classes.noPadding}
                component={Checkbox}
                checkbox
                name={`rules[${idx}].weekdayFilter[${dayIdx}]`}
              />
            </TableCell>
          ))}
        </Hidden>
        <Hidden mdUp>
          <TableCell className={classes.dayFilter}>
            <FormField
              fullWidth
              component={TextField}
              select
              noError
              required
              SelectProps={{
                renderValue: renderDaysValue,
                multiple: true,
              }}
              label=''
              name={`rules[${idx}].weekdayFilter`}
              aria-label='Weekday Filter'
              multiple
              mapValue={(value: boolean[]) =>
                days.filter((d, idx) => value[idx])
              }
              mapOnChangeValue={(value: DayName[]) =>
                days.map((day) => value.includes(day))
              }
            >
              {days.map((day) => (
                <MenuItem value={day} key={day}>
                  {day}
                </MenuItem>
              ))}
            </FormField>
          </TableCell>
        </Hidden>
        <TableCell padding='none'>
          {props.value.rules.length > 1 && (
            <IconButton
              aria-label='Delete rule'
              onClick={() =>
                props.onChange({
                  ...props.value,
                  rules: props.value.rules.filter((r, i) => i !== idx),
                })
              }
            >
              <Trash />
            </IconButton>
          )}
        </TableCell>
      </TableRow>
    )
  }

  return (
    <FormContainer {...formProps} optionalLabels>
      <Grid container spacing={2}>
        <Grid item xs={12} sm={12} md={6} className={classes.tzNote}>
          <Typography color='textSecondary' className={classes.italic}>
            Times and weekdays shown in{' '}
            {displayTZ === 'local' ? 'local time' : displayTZ}.
          </Typography>
        </Grid>
        <Grid item xs={12} sm={12} md={6}>
          {/* Purposefully leaving out of form, as it's only used for converting display times. */}
          <ScheduleTZFilter
            label={(tz) => `Configure in ${tz}`}
            scheduleID={props.scheduleID}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            required
            component={targetType === 'user' ? UserSelect : RotationSelect}
            label={startCase(targetType)}
            disabled={targetDisabled}
            name='target.id'
          />
        </Grid>
        <Grid item xs={12}>
          <Table data-cy='target-rules'>
            <TableHead>
              <TableRow>
                <TableCell className={classes.startEnd}>Start</TableCell>
                <TableCell className={classes.startEnd}>End</TableCell>
                <Hidden smDown>
                  {days.map((d) => (
                    <TableCell key={d} padding='checkbox'>
                      {d.slice(0, 3)}
                    </TableCell>
                  ))}
                </Hidden>
                <Hidden mdUp>
                  <TableCell className={classes.dayFilter}>Days</TableCell>
                </Hidden>
                <TableCell padding='none'>
                  <IconButton
                    aria-label='Add rule'
                    onClick={() =>
                      props.onChange({
                        ...props.value,
                        rules: props.value.rules.concat({
                          start: DateTime.local()
                            .startOf('day')
                            .toUTC()
                            .toISO(),
                          end: DateTime.local()
                            .plus({ day: 1 })
                            .startOf('day')
                            .toUTC()
                            .toISO(),
                          weekdayFilter: Array(days.length).fill(true),
                        }),
                      })
                    }
                  >
                    <Add />
                  </IconButton>
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {props.value.rules.map((r, idx) => renderRuleField(idx))}
            </TableBody>
          </Table>
        </Grid>
      </Grid>
    </FormContainer>
  )
}

export default ScheduleRuleForm
