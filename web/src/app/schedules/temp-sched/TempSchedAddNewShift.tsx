import React, { useEffect, useState } from 'react'
import { Button, Grid } from '@mui/material'
import Accordion from '@mui/material/Accordion'
import AccordionActions from '@mui/material/AccordionActions'
import AccordionSummary from '@mui/material/AccordionSummary'
import AccordionDetails from '@mui/material/AccordionDetails'
import Typography from '@mui/material/Typography'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import ToggleIcon from '@mui/icons-material/CompareArrows'
import _ from 'lodash'
import { dtToDuration, Shift, TempSchedValue } from './sharedUtils'
import { FormContainer, FormField } from '../../forms'
import { DateTime, Interval } from 'luxon'
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
  edit?: boolean
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
  showForm,
  setShowForm,
  shift,
  setShift,
}: AddShiftsStepProps): React.ReactNode {
  const [submitted, setSubmitted] = useState(false)

  const [manualEntry, setManualEntry] = useState(false)
  const { q, zone, isLocalZone } = useScheduleTZ(scheduleID)

  // set start equal to the temporary schedule's start
  // can't this do on mount since the step renderer puts everyone on the DOM at once
  useEffect(() => {
    if (zone === '') return

    setShift({
      start: value.start,
      end: DateTime.fromISO(value.start, { zone }).plus({ hours: 8 }).toISO(),
      userID: '',
      truncated: false,
    })
  }, [value.start, zone])

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
    const diff = end.diff(DateTime.fromISO(shift.start, { zone }))
    setShift({
      userID: '',
      truncated: false,
      start: shift.end,
      end: end.plus(diff).toISO(),
    })
    setSubmitted(false)
  }

  return (
    <FormContainer
      errors={fieldErrors()}
      value={shift}
      onChange={(val: Shift) => setShift(val)}
    >
      <Accordion
        variant='outlined'
        onChange={() => setShowForm(!showForm)}
        expanded={showForm}
      >
        <AccordionSummary
          expandIcon={<ExpandMoreIcon />}
          data-cy='add-shift-expander'
        >
          <Typography
            color='textSecondary'
            variant='button'
            style={{ width: '100%' }}
          >
            ADD SHIFT
          </Typography>
        </AccordionSummary>
        <AccordionDetails data-cy='add-shift-container'>
          <Grid container spacing={2}>
            <Grid item xs={12}>
              <FormField
                fullWidth
                component={UserSelect}
                label='Select a User'
                name='userID'
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
                mapOnChangeValue={(
                  value: string,
                  formValue: TempSchedValue,
                ) => {
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
                disabled={q.loading}
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
                  }
                  timeZone={zone}
                  disabled={q.loading}
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
                  disabled={q.loading}
                  hint={
                    <ClickableText
                      data-cy='toggle-duration-off'
                      onClick={() => setManualEntry(true)}
                      endIcon={<ToggleIcon />}
                    >
                      Configure as date/time
                    </ClickableText>
                  }
                />
              )}
            </Grid>
          </Grid>
        </AccordionDetails>
        <AccordionActions>
          <Button
            data-cy='add-shift'
            color='secondary'
            variant='contained'
            onClick={handleAddShift}
          >
            Add
          </Button>
        </AccordionActions>
      </Accordion>
    </FormContainer>
  )
}
