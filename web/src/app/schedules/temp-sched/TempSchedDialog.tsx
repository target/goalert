import React, { useState, useRef, Suspense, useMemo } from 'react'
import { useMutation, gql, useQuery } from 'urql'
import Checkbox from '@mui/material/Checkbox'
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
import { inferDuration, Shift, TempSchedValue } from './sharedUtils'
import { FormContainer } from '../../forms'
import { parseInterval } from '../../util/shifts'
import { useScheduleTZ } from '../useScheduleTZ'
import TempSchedShiftsList from './TempSchedShiftsList'
import { getCoverageGapItems } from './shiftsListUtil'
import { ensureInterval } from '../timeUtil'
import TempSchedConfirmation from './TempSchedConfirmation'
import { User } from 'web/src/schema'
import TempSchedForm from './TempSchedForm'

const mutation = gql`
  mutation ($input: SetTemporaryScheduleInput!) {
    setTemporarySchedule(input: $input)
  }
`

const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      associatedUsers {
        id
        name
      }
    }
  }
`

function shiftEquals(a: Shift, b: Shift): boolean {
  return a.start === b.start && a.end === b.end && a.userID === b.userID
}

const useStyles = makeStyles((theme: Theme) => ({
  formContainer: {
    height: '100%',
    marginTop: '-16px',
    paddingLeft: '.5rem',
    paddingRight: '.5rem',
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
      paddingLeft: '4rem',
    },
    overflow: 'hidden',
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

export interface DurationValues {
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
  const { q, zone } = useScheduleTZ(scheduleID)
  const now = useMemo(() => DateTime.utc().startOf('minute').toISO(), [])
  const [showForm, setShowForm] = useState(false)
  const [isCustomShiftTimeRange, setIsCustomShiftTimeRange] = useState(false)

  const [{ fetching: fetchingUsers, error: errorUsers, data: dataUsers }] =
    useQuery({
      query,
      variables: {
        id: scheduleID,
      },
    })
  const associatedUsers: Array<User> = dataUsers.schedule.associatedUsers

  const [{ fetching, error }, commit] = useMutation(mutation)

  let defaultShiftDur = {} as DurationValues

  const getDurValues = (dur: Duration): DurationValues => {
    if (!dur) return { ivl: 'days', dur: 1 }
    if (dur.hours < 24 && dur.days < 1)
      return { ivl: 'hours', dur: Math.ceil(dur.hours) }
    if (dur.days < 7) return { ivl: 'days', dur: Math.ceil(dur.days) }

    return { ivl: 'weeks', dur: Math.ceil(dur.weeks) }
  }

  if (edit && _value.shifts.length > 0) {
    // if editing infer shift duration
    defaultShiftDur = getDurValues(inferDuration(_value.shifts))
  } else {
    defaultShiftDur = getDurValues(_value?.shiftDur as Duration)
  }

  const [durValues, setDurValues] = useState<DurationValues>(defaultShiftDur)

  const [value, setValue] = useState<TempSchedValue>({
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
  const [submitAttempt, setSubmitAttempt] = useState(false) // helps with error messaging on step 1
  const [submitSuccess, setSubmitSuccess] = useState(false)

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

    setIsCustomShiftTimeRange(true)
    setShift({
      userID: shift?.userID ?? '',
      truncated: !!shift?.truncated,
      start: nextStart.toISO(),
      end: nextEnd.toISO(),
    })
  }

  const hasCoverageGaps = (() => {
    if (q.loading) return false
    if (!value.shifts || value.shifts.length === 0) return true

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

  const shiftListRef = useRef<HTMLDivElement | null>(null)

  const handleNext = (): void => {
    if (hasCoverageGaps && !allowNoCoverage) {
      setSubmitAttempt(true)
      // Scroll to show gap in coverage error on top of shift list
      if (shiftListRef?.current) {
        shiftListRef.current.scrollIntoView({ behavior: 'smooth' })
      }
    } else {
      setSubmitSuccess(true)
    }
  }
  const handleBack = (): void => {
    setSubmitAttempt(false)
    setSubmitSuccess(false)
  }
  const handleSubmit = (): void => {
    if (hasCoverageGaps && !allowNoCoverage) {
      setSubmitAttempt(true)
      // Scroll to show gap in coverage error on top of shift list
      if (shiftListRef?.current) {
        shiftListRef.current.scrollIntoView({ behavior: 'smooth' })
      }
    } else {
      commit(
        {
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
        { additionalTypenames: ['Schedule'] },
      ).then((result) => {
        if (!result.error) {
          onClose()
        }
      })
    }
  }

  const nonFieldErrs = nonFieldErrors(error).map((e) => ({
    message: e.message,
  }))
  const fieldErrs = fieldErrors(error).map((e) => ({
    message: `${e.field}: ${e.message}`,
  }))
  const noCoverageErrs =
    submitSuccess && hasCoverageGaps && !allowNoCoverage
      ? [new Error('This temporary schedule has gaps in coverage.')]
      : []
  const errs = nonFieldErrs
    .concat(fieldErrs)
    .concat(shiftErrors)
    .concat(noCoverageErrs)

  // if error from loading associated users
  if (errorUsers?.message) {
    errs.concat({ message: errorUsers.message })
  }

  return (
    <FormDialog
      fullScreen
      maxWidth='lg'
      title={edit ? 'Edit a Temporary Schedule' : 'Define a Temporary Schedule'}
      onClose={onClose}
      onSubmit={handleSubmit}
      disableSubmit={edit && !submitSuccess}
      onNext={edit ? handleNext : null}
      disableNext={edit && submitSuccess}
      onBack={edit && submitSuccess ? handleBack : null}
      loading={fetching}
      errors={errs}
      disableBackdropClose
      form={
        <Suspense>
          <FormContainer
            optionalLabels
            disabled={fetchingUsers || fetching}
            value={value}
            onChange={(newValue: TempSchedValue) => {
              setValue({ ...value, ...ensureInterval(value, newValue) })
            }}
          >
            {(edit && submitSuccess && !hasCoverageGaps) ||
            (edit && submitSuccess && hasCoverageGaps && allowNoCoverage) ? (
              <TempSchedConfirmation value={value} scheduleID={scheduleID} />
            ) : (
              <Grid
                container
                className={classes.formContainer}
                justifyContent='space-between'
              >
                {/* left pane */}
                <TempSchedForm
                  scheduleID={scheduleID}
                  associatedUsers={associatedUsers}
                  duration={durValues}
                  setDuration={setDurValues}
                  value={value}
                  setValue={setValue}
                  showForm={showForm}
                  setShowForm={setShowForm}
                  shift={shift}
                  setShift={setShift}
                  isCustomShiftTimeRange={isCustomShiftTimeRange}
                  setIsCustomShiftTimeRange={setIsCustomShiftTimeRange}
                />

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
                    <Typography
                      variant='subtitle1'
                      component='h3'
                      sx={{ fontSize: '1.15rem' }}
                    >
                      Shifts
                    </Typography>

                    {submitAttempt && hasCoverageGaps && (
                      <Alert
                        severity='error'
                        className={classes.noCoverageError}
                      >
                        <AlertTitle>Gaps in coverage</AlertTitle>
                        <FormHelperText>
                          There are gaps in coverage. During these gaps, nobody
                          on the schedule will receive alerts. If you still want
                          to proceed, check the box below and retry.
                        </FormHelperText>
                        <FormControlLabel
                          label='Allow gaps in coverage'
                          labelPlacement='end'
                          control={
                            <Checkbox
                              data-cy='no-coverage-checkbox'
                              checked={allowNoCoverage}
                              onChange={(e) => {
                                setSubmitSuccess(false)
                                setAllowNoCoverage(e.target.checked)
                              }}
                              name='allowCoverageGaps'
                            />
                          }
                        />
                      </Alert>
                    )}

                    {DateTime.fromISO(value.start) >
                      DateTime.utc().minus({ hour: 1 }) || edit ? null : (
                      <Alert
                        severity='warning'
                        className={classes.noCoverageError}
                      >
                        <AlertTitle>Start time occurs in the past</AlertTitle>
                        <FormHelperText>
                          Any shifts or changes made to shifts in the past will
                          be ignored when submitting.
                        </FormHelperText>
                        <FormControlLabel
                          label='Allow gaps in coverage'
                          labelPlacement='end'
                          control={
                            <Checkbox
                              data-cy='no-coverage-checkbox'
                              checked={allowNoCoverage}
                              onChange={(e) =>
                                setAllowNoCoverage(e.target.checked)
                              }
                              name='allowCoverageGaps'
                            />
                          }
                        />
                      </Alert>
                    )}

                    <TempSchedShiftsList
                      scheduleID={scheduleID}
                      value={value.shifts}
                      shiftDur={value.shiftDur}
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
            )}
          </FormContainer>
        </Suspense>
      }
    />
  )
}
