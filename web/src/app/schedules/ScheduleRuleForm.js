import React from 'react'
import p from 'prop-types'
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
import { useURLParam } from '../actions/hooks'
import { UserSelect, RotationSelect } from '../selection'
import { startCase } from 'lodash'
import { Add, Trash } from '../icons'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import Query from '../util/Query'
import { gql } from '@apollo/client'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { ISOTimePicker } from '../util/ISOPickers'
import { DateTime } from 'luxon'

const days = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
]

const renderDaysValue = (value) => {
  const parts = []
  let start = ''
  let last = ''
  let lastIdx = -1

  const flush = () => {
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

const styles = () => {
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
  }
}

const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
    }
  }
`
const useStyles = makeStyles(styles)

export default function ScheduleRuleForm(props) {
  function renderForm() {
    const { targetDisabled, targetType, ...formProps } = props
    const classes = useStyles()
    const [zone] = useURLParam('tz', 'local')

    return (
      <FormContainer {...formProps} optionalLabels>
        <Grid container spacing={2}>
          <Grid item xs={12} sm={12} md={6} className={classes.tzNote}>
            <Typography color='textSecondary' style={{ fontStyle: 'italic' }}>
              Times and weekdays shown in{' '}
              {zone === 'local' ? 'local time' : zone}.
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
              name='targetID'
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

  function renderRuleField(idx) {
    const classes = props.classes
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
          {days.map((day, dayIdx) => (
            <TableCell key={dayIdx} padding='checkbox'>
              <FormField
                noError
                className={props.classes.noPadding}
                component={Checkbox}
                checkbox
                fieldName={`rules[${idx}].weekdayFilter[${dayIdx}]`}
                name={day}
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
              mapValue={(value) => days.filter((d, idx) => value[idx])}
              mapOnChangeValue={(value) =>
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
    <Query
      query={query}
      variables={{ id: props.scheduleID }}
      noPoll
      render={({ data }) => renderForm(data.schedule.timeZone)}
    />
  )
}

ScheduleRuleForm.propTypes = {
  targetType: p.oneOf(['rotation', 'user']).isRequired,
  targetDisabled: p.bool,

  scheduleID: p.string.isRequired,

  value: p.shape({
    targetID: p.string.isRequired,
    rules: p.arrayOf(
      p.shape({
        start: p.string.isRequired,
        end: p.string.isRequired,
        weekdayFilter: p.arrayOf(p.bool).isRequired,
      }),
    ).isRequired,
  }).isRequired,
}
