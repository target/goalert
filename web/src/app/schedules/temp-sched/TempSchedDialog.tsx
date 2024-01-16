import React, { useState, useRef } from 'react'
import { useMutation, gql } from '@apollo/client'
import Checkbox from '@mui/material/Checkbox'
import DialogContentText from '@mui/material/DialogContentText'
import FormControlLabel from '@mui/material/FormControlLabel'
import FormHelperText from '@mui/material/FormHelperText'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import Alert from '@mui/material/Alert'
import AlertTitle from '@mui/material/AlertTitle'
import _ from 'lodash'
import { DateTime, Duration, Interval } from 'luxon'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import {
  contentText,
  inferDuration,
  Shift,
  TempSchedValue,
} from './sharedUtils'
import { FormContainer, FormField } from '../../forms'
import TempSchedAddNewShift from './TempSchedAddNewShift'
import { isISOAfter, parseInterval } from '../../util/shifts'
import { useScheduleTZ } from '../useScheduleTZ'
import TempSchedShiftsList from './TempSchedShiftsList'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { getCoverageGapItems } from './shiftsListUtil'
import { fmtLocal } from '../../util/timeFormat'
import { ensureInterval } from '../timeUtil'
import {
  TextField,
  Select,
  MenuItem,
  InputLabel,
  FormControl,
  SelectChangeEvent,
} from '@mui/material'

const mutation = gql`
  mutation ($input: SetTemporaryScheduleInput!) {
    setTemporarySchedule(input: $input)
  }
`

function shiftEquals(a: Shift, b: Shift): boolean {
  return a.start === b.start && a.end === b.end && a.userID === b.userID
}

const useStyles = makeStyles((theme: Theme) => ({
  contentText,
  avatar: {
    backgroundColor: theme.palette.primary.main,
  },
  formContainer: {
    height: '100%',
  },
  noCoverageError: {
    marginTop: '.5rem',
    marginBottom: '.5rem',
  },
  rightPane: {
    [theme.breakpoints.down('md')]: {
      marginTop: '1rem',
    },
    [theme.breakpoints.up('md')]: {
      paddingLeft: '1rem',
    },
    overflow: 'hidden',
  },
  sticky: {
    position: 'sticky',
    top: 0,
  },
  tzNote: {
    fontStyle: 'italic',
  },
}))

type TempScheduleDialogProps = {
  onClose: () => void
  scheduleID: string
  value: TempSchedValue
  edit?: boolean
}

const clampForward = (nowISO: string, iso: string): string => {
  if (!iso) return ''

  const now = DateTime.fromISO(nowISO)
  const dt = DateTime.fromISO(iso)
  if (dt < now) {
    return now.toISO()
  }
  return iso
}

interface DurationValues {
  dur: number
  ivl: string
}

export default function TempSchedDialog({
  onClose,
  scheduleID,
  value: _value,
  edit = false,
}: TempScheduleDialogProps): JSX.Element {
  const classes = useStyles()
  const { q, zone, isLocalZone } = useScheduleTZ(scheduleID)
  const [now] = useState(DateTime.utc().startOf('minute').toISO())
  const [showForm, setShowForm] = useState(false)

  let defaultShiftDur = {} as DurationValues
  // if editing infer shift duration
  if (!edit) {
    const inferred = inferDuration(_value.shifts)?.toObject()
    defaultShiftDur = {
      ivl: Object.keys(inferred as object)[0],
      dur: Math.floor(Object.values(inferred as object)[0]),
    }
  } else {
    defaultShiftDur = {
      ivl: Object.keys(_value.shiftDur.toObject())[0],
      dur: Object.values(_value.shiftDur.toObject())[0],
    }
  }

  const [durValues, setDurValues] = useState<DurationValues>(defaultShiftDur)

  const [value, setValue] = useState({
    start: clampForward(now, _value.start),
    end: _value.end,
    clearStart: edit ? _value.start : null,
    clearEnd: edit ? _value.end : null,
    shifts: _value.shifts
      .map((s) =>
        _.pick(s, 'start', 'end', 'userID', 'truncated', 'displayStart'),
      )
      .filter((s) => {
        if (DateTime.fromISO(s.end) > DateTime.fromISO(now)) {
          s.displayStart = s.start
          s.start = clampForward(now, s.start)
        }
        return true
      }),
    shiftDur:
      _value.shiftDur ||
      Duration.fromObject({ [durValues.ivl]: durValues.dur }),
  })
  const startDT = DateTime.fromISO(value.start, { zone })
  const [shift, setShift] = useState<Shift>({
    start: startDT.toISO(),
    end: startDT.plus(value.shiftDur).toISO(),
    userID: '',
    truncated: false,
  })
  const [allowNoCoverage, setAllowNoCoverage] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(false)

  function validate(): Error | null {
    if (isISOAfter(value.start, value.end)) {
      return new Error('Start date/time cannot be after end date/time.')
    }
    return null
  }

  const hasInvalidShift = (() => {
    if (q.loading) return false
    const schedInterval = parseInterval(value, zone)
    return value.shifts.some(
      (s) =>
        DateTime.fromISO(s.end) > DateTime.fromISO(now) &&
        !schedInterval.engulfs(parseInterval(s, zone)),
    )
  })()

  const shiftErrors = hasInvalidShift
    ? [
        {
          message:
            'One or more shifts extend beyond the start and/or end of this temporary schedule',
        },
      ]
    : []

  function handleCoverageGapClick(coverageGap: Interval): void {
    if (!showForm) setShowForm(true)

    const nextStart = coverageGap?.start
    const nextEnd = nextStart.plus(value.shiftDur)

    setShift({
      userID: shift?.userID ?? '',
      truncated: !!shift?.truncated,
      start: nextStart.toISO(),
      end: nextEnd.toISO(),
    })
  }

  const hasCoverageGaps = (() => {
    if (q.loading) return false
    const schedInterval = parseInterval(value, zone)
    return (
      getCoverageGapItems(
        schedInterval,
        value.shiftDur,
        value.shifts,
        zone,
        handleCoverageGapClick,
      ).length > 0
    )
  })()

  const [submit, { loading, error }] = useMutation(mutation, {
    onCompleted: () => onClose(),
    variables: {
      input: {
        start: value.start,
        end: value.end,
        clearStart: value.clearStart,
        clearEnd: value.clearEnd,
        shifts: value.shifts
          .map((s) => _.pick(s, 'start', 'end', 'userID'))
          .filter((s) => {
            // clamp/filter out shifts that are in the past
            if (DateTime.fromISO(s.end) <= DateTime.fromISO(now)) {
              return false
            }

            s.start = clampForward(now, s.start)
            return true
          }),
        scheduleID,
      },
    },
  })

  const shiftListRef = useRef<HTMLDivElement | null>(null)

  const handleSubmit = (): void => {
    setHasSubmitted(true)

    if (hasCoverageGaps && !allowNoCoverage) {
      // Scroll to show gap in coverage error on top of shift list
      if (shiftListRef?.current) {
        shiftListRef.current.scrollIntoView({ behavior: 'smooth' })
      }
      return
    }

    submit()
  }

  const nonFieldErrs = nonFieldErrors(error).map((e) => ({
    message: e.message,
  }))
  const fieldErrs = fieldErrors(error).map((e) => ({
    message: `${e.field}: ${e.message}`,
  }))
  const noCoverageErrs =
    hasSubmitted && hasCoverageGaps && !allowNoCoverage
      ? [new Error('This temporary schedule has gaps in coverage.')]
      : []
  const errs = nonFieldErrs
    .concat(fieldErrs)
    .concat(shiftErrors)
    .concat(noCoverageErrs)

  return (
    <FormDialog
      fullHeight
      maxWidth='lg'
      title='Define a Temporary Schedule'
      onClose={onClose}
      loading={loading}
      errors={errs}
      disableBackdropClose
      notices={
        DateTime.fromISO(value.start) > DateTime.utc().minus({ hour: 1 }) ||
        edit
          ? []
          : [
              {
                type: 'WARNING',
                message: 'Start time occurs in the past',
                details:
                  'Any shifts or changes made to shifts in the past will be ignored when submitting.',
              },
            ]
      }
      form={
        <FormContainer
          optionalLabels
          disabled={loading}
          value={value}
          onChange={(newValue: TempSchedValue) => {
            setValue({
              ...value,
              ...ensureInterval(value, newValue),
            })
          }}
        >
          <Grid
            container
            className={classes.formContainer}
            justifyContent='space-between'
          >
            {/* left pane */}
            <Grid
              item
              xs={12}
              md={6}
              container
              alignContent='flex-start'
              spacing={2}
            >
              <Grid item xs={12}>
                <DialogContentText className={classes.contentText}>
                  The schedule will be exactly as configured here for the entire
                  duration (ignoring all assignments and overrides).
                </DialogContentText>
              </Grid>

              <Grid item xs={12}>
                <Typography color='textSecondary' className={classes.tzNote}>
                  Times shown in schedule timezone ({zone})
                </Typography>
              </Grid>

              <Grid item xs={12} md={6}>
                <FormField
                  fullWidth
                  component={ISODateTimePicker}
                  required
                  name='start'
                  label='Schedule Start'
                  min={now}
                  max={DateTime.fromISO(now, { zone })
                    .plus({ year: 1 })
                    .toISO()}
                  softMax={value.end}
                  softMaxLabel='Must be before end time.'
                  softMin={DateTime.fromISO(value.end)
                    .plus({ month: -3 })
                    .toISO()}
                  softMinLabel='Must be within 3 months of end time.'
                  validate={() => validate()}
                  timeZone={zone}
                  disabled={q.loading}
                  hint={isLocalZone ? '' : fmtLocal(value.start)}
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <FormField
                  fullWidth
                  component={ISODateTimePicker}
                  required
                  name='end'
                  label='Schedule End'
                  min={now}
                  softMin={value.start}
                  softMinLabel='Must be after start time.'
                  softMax={DateTime.fromISO(value.start)
                    .plus({ month: 3 })
                    .toISO()}
                  softMaxLabel='Must be within 3 months of start time.'
                  validate={() => validate()}
                  timeZone={zone}
                  disabled={q.loading}
                  hint={isLocalZone ? '' : fmtLocal(value.end)}
                />
              </Grid>

              <FormContainer
                value={durValues}
                onChange={(newValue: DurationValues) => {
                  if (newValue.dur) {
                    setDurValues({ ...durValues, ...newValue })
                    setValue({
                      ...value,
                      shiftDur: Duration.fromObject({
                        [newValue.ivl]: newValue.dur,
                      }),
                    })
                  }
                }}
              >
                <Grid item xs={12} md={6}>
                  <FormField
                    fullWidth
                    component={TextField}
                    required
                    type='number'
                    name='dur'
                    label='Shift Duration'
                    min={1}
                    validate={() => validate()}
                    disabled={q.loading}
                  />
                </Grid>
                <Grid item xs={12} md={6}>
                  <FormControl sx={{ width: '100%' }}>
                    <InputLabel>Shift Interval</InputLabel>
                    <Select
                      fullWidth
                      required
                      name='ivl'
                      value={durValues.ivl}
                      onChange={(e: SelectChangeEvent<string>) => {
                        setDurValues({ ...durValues, ivl: e.target.value })
                        setValue({
                          ...value,
                          shiftDur: Duration.fromObject({
                            [e.target.value]: durValues.dur,
                          }),
                        })
                      }}
                    >
                      <MenuItem value='hours'>Hour</MenuItem>
                      <MenuItem value='days'>Day</MenuItem>
                      <MenuItem value='weeks'>Week</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>
              </FormContainer>

              <Grid item xs={12} className={classes.sticky}>
                <TempSchedAddNewShift
                  value={value}
                  onChange={(shifts: Shift[]) => setValue({ ...value, shifts })}
                  scheduleID={scheduleID}
                  edit={edit}
                  showForm={showForm}
                  setShowForm={setShowForm}
                  shift={shift}
                  setShift={setShift}
                />
              </Grid>
            </Grid>

            {/* right pane */}
            <Grid
              item
              xs={12}
              md={6}
              container
              spacing={2}
              className={classes.rightPane}
            >
              <Grid item xs={12} ref={shiftListRef}>
                <Typography variant='subtitle1' component='h3'>
                  Shifts
                </Typography>

                {hasSubmitted && hasCoverageGaps && (
                  <Alert severity='error' className={classes.noCoverageError}>
                    <AlertTitle>Gaps in coverage</AlertTitle>
                    <FormHelperText>
                      There are gaps in coverage. During these gaps, nobody on
                      the schedule will receive alerts. If you still want to
                      proceed, check the box below and retry.
                    </FormHelperText>
                    <FormControlLabel
                      label='Allow gaps in coverage'
                      labelPlacement='end'
                      control={
                        <Checkbox
                          data-cy='no-coverage-checkbox'
                          checked={allowNoCoverage}
                          onChange={(e) => setAllowNoCoverage(e.target.checked)}
                          name='allowCoverageGaps'
                        />
                      }
                    />
                  </Alert>
                )}

                <TempSchedShiftsList
                  scheduleID={scheduleID}
                  shiftDur={value.shiftDur}
                  value={value.shifts}
                  start={value.start}
                  end={value.end}
                  onRemove={(shift: Shift) => {
                    setValue({
                      ...value,
                      shifts: value.shifts.filter(
                        (s) => !shiftEquals(shift, s),
                      ),
                    })
                  }}
                  edit={edit}
                  handleCoverageGapClick={handleCoverageGapClick}
                />
              </Grid>
            </Grid>
          </Grid>
        </FormContainer>
      }
      onSubmit={handleSubmit}
    />
  )
}
