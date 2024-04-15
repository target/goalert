import React, { useEffect, useState } from 'react'
import { Button, Checkbox, FormControlLabel, Grid } from '@mui/material'
import Typography from '@mui/material/Typography'
import ToggleIcon from '@mui/icons-material/CompareArrows'
import _ from 'lodash'
import { dtToDuration, Shift, TempSchedValue } from './sharedUtils'
import { FormContainer, FormField } from '../../forms'
import { DateTime, Duration, Interval } from 'luxon'
import { FieldError } from '../../util/errutil'
import { isISOAfter } from '../../util/shifts'
import { useScheduleTZ } from '../useScheduleTZ'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { UserSelect } from '../../selection'
import ClickableText from '../../util/ClickableText'
import NumberField from '../../util/NumberField'
import { fmtLocal } from '../../util/timeFormat'

type AddShiftsStepProps = {
  value: TempSchedValue
  onChange: (newValue: Shift[]) => void

  scheduleID: string
  showForm: boolean
  setShowForm: (showForm: boolean) => void
  shift: Shift | null
  setShift: (shift: Shift) => void
}

type DTShift = {
  userID: string
  span: Interval
}

function shiftsToDT(shifts: Shift[]): DTShift[] {
  return shifts.map((s) => ({
    userID: s.userID,
    span: Interval.fromDateTimes(
      DateTime.fromISO(s.start),
      DateTime.fromISO(s.end),
    ),
  }))
}

function DTToShifts(shifts: DTShift[]): Shift[] {
  return shifts.map((s) => ({
    userID: s.userID,
    start: s.span.start.toISO(),
    end: s.span.end.toISO(),
    truncated: false,
  }))
}

// mergeShifts will take the incoming shifts and merge them with
// the shifts stored in value. Using Luxon's Interval, overlaps
// and edge cases when merging are handled for us.
function mergeShifts(_shifts: Shift[]): Shift[] {
  const byUser = _.groupBy(shiftsToDT(_shifts), 'userID')

  return DTToShifts(
    _.flatten(
      _.values(
        _.mapValues(byUser, (shifts, userID) => {
          return Interval.merge(_.map(shifts, 'span')).map((span) => ({
            userID,
            span,
          }))
        }),
      ),
    ),
  )
}

export default function TempSchedAddNewShift({
  scheduleID,
  onChange,
  value,
  shift,
  setShift,
}: AddShiftsStepProps): JSX.Element {
  const [submitted, setSubmitted] = useState(false)

  const [custom, setCustom] = useState(false)
  const [manualEntry, setManualEntry] = useState(true)
  const { q, zone, isLocalZone } = useScheduleTZ(scheduleID)

  // set start equal to the temporary schedule's start
  // can't this do on mount since the step renderer puts everyone on the DOM at once
  useEffect(() => {
    if (zone === '') return

    setShift({
      start: value.start,
      end: DateTime.fromISO(value.start, { zone })
        .plus(value.shiftDur as Duration)
        .toISO(),
      userID: '',
      truncated: false,
    })
  }, [value.start, zone, value.shiftDur])

  // fieldErrors handles errors manually through the client
  // as this step form is nested inside the greater form
  // that makes the network request.
  function fieldErrors(s = submitted): FieldError[] {
    const result: FieldError[] = []
    const requiredMsg = 'this field is required'
    const add = (field: string, message: string): void => {
      result.push({ field, message } as FieldError)
    }

    if (!shift) return result
    if (s) {
      if (!shift.userID) add('userID', requiredMsg)
      if (!shift.start) add('start', requiredMsg)
      if (!shift.end) add('end', requiredMsg)
    }

    if (!isISOAfter(shift.end, shift.start)) {
      add('end', 'must be after shift start time')
      add('start', 'must be before shift end time')
    }

    return result
  }

  function handleAddShift(): void {
    if (fieldErrors(true).length) {
      setSubmitted(true)
      return
    }
    if (!shift) return // ts sanity check

    onChange(mergeShifts(value.shifts.concat(shift)))
    const end = DateTime.fromISO(shift.end, { zone })
    setShift({
      userID: '',
      truncated: false,
      start: shift.end,
      end: end.plus(value.shiftDur as Duration).toISO(),
    })
    setCustom(false)
    setSubmitted(false)
  }

  return (
    <FormContainer
      errors={fieldErrors()}
      value={shift}
      onChange={(val: Shift) => setShift(val)}
    >
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Typography color='textSecondary'>Add Shift</Typography>
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={UserSelect}
            label='Select a User'
            name='userID'
          />
        </Grid>
        <Grid item xs={12}>
          <FormControlLabel
            control={<Checkbox checked={custom} data-cy='toggle-custom' />}
            label={
              <Typography color='textSecondary' sx={{ fontStyle: 'italic' }}>
                Configure custom shift
              </Typography>
            }
            onChange={() => setCustom(!custom)}
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            label='Shift Start'
            name='shift-start'
            fieldName='start'
            min={value.start}
            max={DateTime.fromISO(value.end, { zone })
              .plus({ year: 1 })
              .toISO()}
            mapOnChangeValue={(value: string, formValue: TempSchedValue) => {
              if (!manualEntry) {
                const diff = DateTime.fromISO(value, { zone }).diff(
                  DateTime.fromISO(formValue.start, { zone }),
                )
                formValue.end = DateTime.fromISO(formValue.end, { zone })
                  .plus(diff)
                  .toISO()
              }
              return value
            }}
            timeZone={zone}
            disabled={q.loading || !custom}
            hint={isLocalZone ? '' : fmtLocal(value?.start)}
          />
        </Grid>
        <Grid item xs={6}>
          {manualEntry ? (
            <FormField
              fullWidth
              component={ISODateTimePicker}
              label='Shift End'
              name='shift-end'
              fieldName='end'
              min={value.start}
              max={DateTime.fromISO(value.end, { zone })
                .plus({ year: 1 })
                .toISO()}
              hint={
                custom ? (
                  <React.Fragment>
                    {!isLocalZone && fmtLocal(value?.end)}
                    <div>
                      <ClickableText
                        data-cy='toggle-duration-on'
                        onClick={() => setManualEntry(false)}
                        endIcon={<ToggleIcon />}
                      >
                        Configure as duration
                      </ClickableText>
                    </div>
                  </React.Fragment>
                ) : null
              }
              timeZone={zone}
              disabled={q.loading || !custom}
            />
          ) : (
            <FormField
              fullWidth
              component={NumberField}
              label='Shift Duration (hours)'
              name='shift-end'
              fieldName='end'
              float
              // value held in form input
              mapValue={(nextVal: string, formValue: TempSchedValue) => {
                const nextValDT = DateTime.fromISO(nextVal, { zone })
                const formValDT = DateTime.fromISO(formValue?.start ?? '', {
                  zone,
                })
                const duration = dtToDuration(formValDT, nextValDT)
                return duration === -1 ? '' : duration.toString()
              }}
              // value held in state
              mapOnChangeValue={(
                nextVal: string,
                formValue: TempSchedValue,
              ) => {
                if (!nextVal) return ''
                return DateTime.fromISO(formValue.start, { zone })
                  .plus({ hours: parseFloat(nextVal) })
                  .toISO()
              }}
              step='any'
              min={0}
              disabled={q.loading || !custom}
              hint={
                custom ? (
                  <ClickableText
                    data-cy='toggle-duration-off'
                    onClick={() => setManualEntry(true)}
                    endIcon={<ToggleIcon />}
                  >
                    Configure as date/time
                  </ClickableText>
                ) : null
              }
            />
          )}
        </Grid>
        <Grid item xs={12} container justifyContent='flex-end'>
          <Button
            data-cy='add-shift'
            color='secondary'
            variant='contained'
            onClick={handleAddShift}
          >
            Add
          </Button>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
