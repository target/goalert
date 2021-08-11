import React, { useState, ReactNode, useEffect } from 'react'
import { useMutation, gql } from '@apollo/client'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { Shift, Value } from './sharedUtils'
import _ from 'lodash'
import { FormContainer } from '../../forms'
import { virtualize } from 'react-swipeable-views-utils'
import SwipeableViews from 'react-swipeable-views'
import TempSchedAddShiftsStep from './TempSchedAddShiftsStep'
import TempSchedTimesStep from './TempSchedTimesStep'
import { parseInterval } from '../../util/shifts'
import { DateTime } from 'luxon'
import { getNextWeekday } from '../../util/luxon-helpers'
import { useScheduleTZ } from './hooks'
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

export default function TempSchedDialog({
  onClose,
  scheduleID,
  value: _value,
}: TempScheduleDialogProps): JSX.Element {
  const edit = Boolean(_value)
  const { q, zone } = useScheduleTZ(scheduleID)
  const [step, setStep] = useState(edit ? 1 : 0) // edit starting on 2nd step
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
  const fieldErrs = fieldErrors(error).map((e) => ({
    message: `${e.field}: ${e.message}`,
  }))
  const errs = nonFieldErrs.concat(fieldErrs).concat(shiftErrors)

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
      onSubmit={() => submit()}
      onNext={step === 1 ? null : () => setStep(step + 1)}
      onBack={(edit ? step === 1 : step === 0) ? null : () => setStep(step - 1)}
    />
  )
}
