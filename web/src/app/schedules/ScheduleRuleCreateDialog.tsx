import React, { useState } from 'react'
import { useQuery, useMutation, gql } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import ScheduleRuleForm from './ScheduleRuleForm'
import { nonFieldErrors } from '../util/errutil'
import { startCase } from 'lodash'
import { DateTime } from 'luxon'
import { isoToGQLClockTime } from './util'
import { TargetType } from '../../schema'

const mutation = gql`
  mutation ($input: ScheduleTargetInput!) {
    updateScheduleTarget(input: $input)
  }
`
const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
    }
  }
`

interface ScheduleRuleCreateDialogProps {
  scheduleID: string
  targetType: TargetType
  onClose: () => void
}

export default function ScheduleRuleCreateDialog(
  props: ScheduleRuleCreateDialogProps,
): JSX.Element {
  const { scheduleID, targetType, onClose } = props

  const [{ data, fetching }] = useQuery({
    query,
    variables: { id: scheduleID },
  })

  const [value, setValue] = useState({
    targetID: '',
    rules: [
      {
        start: DateTime.local({ zone: data.schedule.timeZone })
          .startOf('day')
          .toISO(),
        end: DateTime.local({ zone: data.schedule.timeZone })
          .plus({ day: 1 })
          .startOf('day')
          .toISO(),
        weekdayFilter: [true, true, true, true, true, true, true],
      },
    ],
  })

  const [{ error }, mutate] = useMutation(mutation)

  return (
    <FormDialog
      onClose={onClose}
      title={`Add ${startCase(targetType)} to Schedule`}
      errors={nonFieldErrors(error)}
      maxWidth='md'
      loading={!data && fetching}
      onSubmit={() =>
        mutate(
          {
            input: {
              target: {
                type: targetType,
                id: value.targetID,
              },
              scheduleID,

              rules: value.rules.map((r) => ({
                ...r,
                start: isoToGQLClockTime(r.start, data.schedule.timeZone),
                end: isoToGQLClockTime(r.end, data.schedule.timeZone),
              })),
            },
          },
          { additionalTypenames: ['Schedule'] },
        ).then(() => onClose())
      }
      form={
        <ScheduleRuleForm
          targetType={targetType}
          scheduleID={scheduleID}
          targetDisabled={!data && fetching}
          value={value}
          onChange={setValue}
        />
      }
    />
  )
}
