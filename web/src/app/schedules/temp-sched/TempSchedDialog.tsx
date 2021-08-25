import React, { useState, ReactNode, useEffect } from 'react'
import { useMutation, gql, ApolloError } from '@apollo/client'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { Shift, Value } from './sharedUtils'
import _ from 'lodash'
import { FormContainer, FormField } from '../../forms'
import { virtualize } from 'react-swipeable-views-utils'
import SwipeableViews from 'react-swipeable-views'
import TempSchedAddShiftsStep from './TempSchedAddShiftsStep'
import TempSchedTimesStep from './TempSchedTimesStep'
import { parseInterval } from '../../util/shifts'
import { DateTime } from 'luxon'
import { getNextWeekday } from '../../util/luxon-helpers'
import { useScheduleTZ } from './hooks'
import {
  Box,
  Checkbox,
  DialogContent,
  FormControlLabel,
  Grid,
  makeStyles,
  Typography,
  Zoom,
} from '@material-ui/core'
import { styles as globalStyles } from '../../styles/materialStyles'
import Error from '@material-ui/icons/Error'

// allows changing the index programatically
const VirtualizeAnimatedViews = virtualize(SwipeableViews)

const mutation = gql`
  mutation ($input: SetTemporaryScheduleInput!) {
    setTemporarySchedule(input: $input)
  }
`

type TempScheduleDialogProps = {
  onClose: () => void
  scheduleID: string
  value?: Value
}

const useStyles = makeStyles((theme) => {
  const { error } = globalStyles(theme)

  return {
    error,
    errorContainer: {
      flexGrow: 0,
      overflowY: 'visible',
    },
  }
})

export default function TempSchedDialog({
  onClose,
  scheduleID,
  value: _value,
}: TempScheduleDialogProps): JSX.Element {
  const classes = useStyles()
  const edit = Boolean(_value)
  const { q, zone } = useScheduleTZ(scheduleID)
  const [step, setStep] = useState(edit ? 1 : 1) // edit starting on 2nd step
  const [value, setValue] = useState({
    start: _value?.start ?? '',
    end: _value?.end ?? '',
    shifts: (_value?.shifts ?? []).map((s) =>
      _.pick(s, 'start', 'end', 'userID'),
    ),
  })

  useEffect(() => {
    // set default start, end times when zone is ready
    if (!value.start && !value.end && !q.loading && zone) {
      const nextSunday = getNextWeekday(7, DateTime.now(), zone)
      const followingSunday = nextSunday.plus({ week: 1 })
      setValue({
        ...value,
        start: nextSunday.toISO(),
        end: followingSunday.toISO(),
      })
    }
  }, [q.loading, zone])

  const schedInterval = parseInterval(value)
  const hasInvalidShift = value.shifts.some(
    (s) => !schedInterval.engulfs(parseInterval(s)),
  )

  const shiftErrors = hasInvalidShift
    ? [
        {
          message:
            'One or more shifts extend beyond the start and/or end of this temporary schedule',
          nonSubmit: step !== 1,
        },
      ]
    : []

  const [submit, { loading, error }] = useMutation(mutation, {
    onCompleted: () => onClose(),
    variables: {
      input: {
        ...value,
        scheduleID,
      },
    },
  })

  const [isAllowingNoCoverage, setIsAllowingNoCoverage] = useState(false)
  const [isNoCoverage, setIsNoCoverage] = useState(false)

  const handleSubmit = (): void => {
    if (!isAllowingNoCoverage) {
      setIsNoCoverage(true)
      return
    }

    submit()
  }

  type SlideRenderer = {
    index: number
    key: number
  }
  function renderSlide({ index, key }: SlideRenderer): ReactNode {
    if (index === 0) {
      return (
        <TempSchedTimesStep
          key={key}
          scheduleID={scheduleID}
          value={value}
          edit={edit}
        />
      )
    }

    if (index === 1) {
      return (
        <TempSchedAddShiftsStep
          key={key}
          value={value.shifts}
          onChange={(shifts: Shift[]) => setValue({ ...value, shifts })}
          scheduleID={scheduleID}
          start={value.start}
          end={value.end}
          edit={edit}
        />
      )
    }

    // fallback empty div
    return <div />
  }

  const nonFieldErrs = nonFieldErrors(error).map((e) => ({
    message: e.message,
  }))

  const noCoverageErrs = isNoCoverage
    ? [
        {
          render: (
            <DialogContent className={classes.errorContainer}>
              <Zoom in>
                <Box width={1 / 3}>
                  <Typography
                    component='div'
                    variant='subtitle1'
                    style={{ display: 'flex' }}
                  >
                    <Error className={classes.error} />
                    &nbsp;
                    <div className={classes.error}>
                      <div>There are gaps in coverage.</div>
                    </div>
                  </Typography>
                  <Typography
                    component='p'
                    variant='caption'
                    style={{ display: 'flex' }}
                    className={classes.error}
                  >
                    This means there could be time periods where your alerts are
                    sent to your entire team instead of a designated on-call
                    person.
                  </Typography>
                  <Typography
                    component='div'
                    variant='subtitle1'
                    style={{ display: 'flex' }}
                    className={classes.error}
                  >
                    <FormControlLabel
                      control={
                        <Checkbox
                          onChange={(e) =>
                            setIsAllowingNoCoverage(e.target.checked)
                          }
                          value={isAllowingNoCoverage}
                        />
                      }
                      label='I would like to allow gaps in coverage'
                      labelPlacement='end'
                    />
                  </Typography>
                </Box>
              </Zoom>
            </DialogContent>
          ),
          message: 'You have shifts with no coverage.',
        },
      ]
    : []

  const fieldErrs = fieldErrors(error).map((e) => ({
    message: `${e.field}: ${e.message}`,
  }))
  const errs = nonFieldErrs
    .concat(fieldErrs)
    .concat(shiftErrors)
    .concat(noCoverageErrs)

  return (
    <FormDialog
      fullScreen
      disableGutters
      title='Define a Temporary Schedule'
      onClose={onClose}
      loading={loading}
      errors={errs}
      notices={
        !value.start ||
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
          onChange={(newValue: Value) => setValue(newValue)}
        >
          <VirtualizeAnimatedViews
            index={step}
            onChangeIndex={(i: number) => {
              if (i < 0 || i > 1) return
              if (edit) {
                setStep(1)
                return
              }
              setStep(i)
            }}
            slideCount={2}
            slideRenderer={renderSlide}
            disabled // disables slides from changing outside of action buttons
            containerStyle={{ height: '100%' }}
            style={{ height: '100%' }}
          />
        </FormContainer>
      }
      onSubmit={handleSubmit}
      onNext={step === 1 ? null : () => setStep(step + 1)}
      onBack={(edit ? step === 1 : step === 0) ? null : () => setStep(step - 1)}
    />
  )
}
