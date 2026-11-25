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
  // below sm there is no room for start, end and days side by side
  const isNarrow = useIsWidthDown('sm')

  const Spacer = (): ReactNode =>
    isLocalZone ? null : <FormHelperText>&nbsp;</FormHelperText>

  function startField(idx: number, label = ''): ReactNode {
    return (
      <FormField
        fullWidth
        noError
        component={ISOTimePicker}
        required
        label={label}
        name={`rules[${idx}].start`}
        disabled={!zone || disabled}
        timeZone={zone}
        hint={isLocalZone ? '' : fmtLocal(value.rules[idx].start)}
      />
    )
  }

  function endField(idx: number, label = ''): ReactNode {
    return (
      <FormField
        fullWidth
        noError
        component={ISOTimePicker}
        required
        label={label}
        name={`rules[${idx}].end`}
        disabled={!zone || disabled}
        timeZone={zone}
        hint={isLocalZone ? '' : fmtLocal(value.rules[idx].end)}
      />
    )
  }

  function daysField(idx: number, label = ''): ReactNode {
    return (
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
        label={label}
        name={`rules[${idx}].weekdayFilter`}
        aria-label='Weekday Filter'
        multiple
        mapValue={(value: string[]) => days.filter((d, idx) => value[idx])}
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
    )
  }

  function deleteButton(idx: number): ReactNode {
    return (
      <React.Fragment>
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
      </React.Fragment>
    )
  }

  const addButton = (
    <IconButton
      aria-label='Add rule'
      disabled={disabled}
      onClick={() =>
        onChange({
          ...value,
          rules: value.rules.concat({
            start: DateTime.local().startOf('day').toUTC().toISO(),
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
  )

  // narrow layout: one table row per rule (kept for tests and a11y),
  // with the fields stacked inside a single cell
  function renderNarrowRuleField(idx: number): ReactNode {
    return (
      <TableRow key={idx}>
        <TableCell
          colSpan={2}
          className={classNames(classes.noPadding, classes.noBorder)}
        >
          <Grid container spacing={1} alignItems='center'>
            <Grid item xs={6}>
              {startField(idx, 'Start')}
            </Grid>
            <Grid item xs={6}>
              {endField(idx, 'End')}
            </Grid>
            <Grid item xs>
              {daysField(idx, 'Days')}
            </Grid>
            {props.value.rules.length > 1 && (
              <Grid item xs='auto'>
                {deleteButton(idx)}
              </Grid>
            )}
          </Grid>
        </TableCell>
      </TableRow>
    )
  }

  function renderRuleField(idx: number): ReactNode {
    return (
      <TableRow key={idx}>
        <TableCell className={classNames(classes.startEnd, classes.noBorder)}>
          {startField(idx)}
        </TableCell>
        <TableCell className={classNames(classes.startEnd, classes.noBorder)}>
          {endField(idx)}
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
            {daysField(idx)}
            <Spacer />
          </TableCell>
        </Hidden>
        <TableCell padding='none' className={classes.noBorder}>
          {deleteButton(idx)}
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
              {isNarrow ? (
                <TableRow>
                  <TableCell
                    className={classNames(classes.noPadding, classes.noBorder)}
                  >
                    Rules
                  </TableCell>
                  <TableCell
                    padding='none'
                    align='right'
                    className={classes.noBorder}
                  >
                    {addButton}
                  </TableCell>
                </TableRow>
              ) : (
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
                      className={classNames(
                        classes.dayFilter,
                        classes.noBorder,
                      )}
                    >
                      Days
                    </TableCell>
                  </Hidden>
                  <TableCell
                    padding='none'
                    className={classNames({ [classes.noBorder]: isMobile })}
                  >
                    {addButton}
                  </TableCell>
                </TableRow>
              )}
            </TableHead>
            <TableBody>
              {props.value.rules.map((r, idx) =>
                isNarrow ? renderNarrowRuleField(idx) : renderRuleField(idx),
              )}
            </TableBody>
          </Table>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
