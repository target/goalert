import React, { ReactNode } from 'react'
import classNames from 'classnames'
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
  FormHelperText,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { UserSelect, RotationSelect } from '../selection'
import { startCase } from 'lodash'
import { Add, Trash } from '../icons'
import { ISOTimePicker } from '../util/ISOPickers'
import { DateTime } from 'luxon'
import { useScheduleTZ } from './useScheduleTZ'
import { fmtLocal } from '../util/timeFormat'
import { useIsWidthDown } from '../util/useWidth'
import { TargetType } from '../../schema'
import { FieldError } from '../util/errutil'

const days = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
]

const renderDaysValue = (value: string): string => {
  const parts = [] as string[]
  let start = ''
  let last = ''
  let lastIdx = -1

  const flush = (): void => {
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

const useStyles = makeStyles({
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
  noBorder: {
    border: 'none',
  },
  table: {
    borderCollapse: 'separate',
    borderSpacing: '0px 5px',
  },
})

export type ScheduleRuleFormValue = {
  targetID: string
  rules: {
    start: string
    end: string
    weekdayFilter: boolean[]
  }[]
}

interface ScheduleRuleFormProps {
  targetType: TargetType
  targetDisabled?: boolean // can't change target when editing
  disabled?: boolean

  scheduleID: string
  value: ScheduleRuleFormValue
  onChange: (value: ScheduleRuleFormValue) => void
  errors?: FieldError[]
}

export default function ScheduleRuleForm(
  props: ScheduleRuleFormProps,
): ReactNode {
  const {
    value,
    scheduleID,
    onChange,
    disabled,
    targetDisabled,
    targetType,
    ...formProps
  } = props
  const classes = useStyles()
  const { zone, isLocalZone } = useScheduleTZ(scheduleID)
  const isMobile = useIsWidthDown('md')

  const Spacer = (): ReactNode =>
    isLocalZone ? null : <FormHelperText>&nbsp;</FormHelperText>

  function renderRuleField(idx: number): ReactNode {
    return (
      <TableRow key={idx}>
        <TableCell className={classNames(classes.startEnd, classes.noBorder)}>
          <FormField
            fullWidth
            noError
            component={ISOTimePicker}
            required
            label=''
            name={`rules[${idx}].start`}
            disabled={!zone || disabled}
            timeZone={zone}
            hint={isLocalZone ? '' : fmtLocal(value.rules[idx].start)}
          />
        </TableCell>
        <TableCell className={classNames(classes.startEnd, classes.noBorder)}>
          <FormField
            fullWidth
            noError
            component={ISOTimePicker}
            required
            label=''
            name={`rules[${idx}].end`}
            disabled={!zone || disabled}
            timeZone={zone}
            hint={isLocalZone ? '' : fmtLocal(value.rules[idx].end)}
          />
        </TableCell>
        <Hidden mdDown>
          {days.map((day, dayIdx) => (
            <TableCell
              key={dayIdx}
              padding='checkbox'
              className={classes.noBorder}
            >
              <FormField
                noError
                className={classes.noPadding}
                component={Checkbox}
                checkbox
                disabled={disabled}
                fieldName={`rules[${idx}].weekdayFilter[${dayIdx}]`}
                name={day}
              />
              <Spacer />
            </TableCell>
          ))}
        </Hidden>
        <Hidden mdUp>
          <TableCell
            className={classNames(classes.dayFilter, classes.noBorder)}
          >
            <FormField
              fullWidth
              component={TextField}
              select
              noError
              required
              disabled={disabled}
              SelectProps={{
                renderValue: renderDaysValue,
                multiple: true,
              }}
              label=''
              name={`rules[${idx}].weekdayFilter`}
              aria-label='Weekday Filter'
              multiple
              mapValue={(value: string[]) =>
                days.filter((d, idx) => value[idx])
              }
              mapOnChangeValue={(value: string[]) =>
                days.map((day) => value.includes(day))
              }
            >
              {days.map((day) => (
                <MenuItem value={day} key={day}>
                  {day}
                </MenuItem>
              ))}
            </FormField>
            <Spacer />
          </TableCell>
        </Hidden>
        <TableCell padding='none' className={classes.noBorder}>
          {props.value.rules.length > 1 && (
            <IconButton
              aria-label='Delete rule'
              disabled={disabled}
              onClick={() =>
                onChange({
                  ...value,
                  rules: value.rules.filter((r, i) => i !== idx),
                })
              }
              size='large'
            >
              <Trash />
            </IconButton>
          )}
          <Spacer />
        </TableCell>
      </TableRow>
    )
  }

  return (
    <FormContainer
      {...formProps}
      value={value}
      onChange={onChange}
      optionalLabels
    >
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            required
            component={targetType === 'user' ? UserSelect : RotationSelect}
            label={startCase(targetType)}
            disabled={disabled || targetDisabled}
            name='targetID'
          />
        </Grid>
        <Grid item xs={12}>
          <Typography color='textSecondary' sx={{ fontStyle: 'italic' }}>
            Times shown in schedule timezone ({zone || '...'})
          </Typography>
        </Grid>
        <Grid item xs={12} style={{ paddingTop: 0 }}>
          <Table data-cy='target-rules' className={classes.table}>
            <TableHead>
              <TableRow>
                <TableCell
                  className={classNames(classes.startEnd, classes.noBorder)}
                >
                  Start
                </TableCell>
                <TableCell
                  className={classNames(classes.startEnd, classes.noBorder)}
                >
                  End
                </TableCell>
                <Hidden mdDown>
                  {days.map((d) => (
                    <TableCell key={d} padding='checkbox'>
                      {d.slice(0, 3)}
                    </TableCell>
                  ))}
                </Hidden>
                <Hidden mdUp>
                  <TableCell
                    className={classNames(classes.dayFilter, classes.noBorder)}
                  >
                    Days
                  </TableCell>
                </Hidden>
                <TableCell
                  padding='none'
                  className={classNames({ [classes.noBorder]: isMobile })}
                >
                  <IconButton
                    aria-label='Add rule'
                    disabled={disabled}
                    onClick={() =>
                      onChange({
                        ...value,
                        rules: value.rules.concat({
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
                    size='large'
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
