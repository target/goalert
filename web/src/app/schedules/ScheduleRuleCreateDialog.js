import React, { useState } from 'react'
import p from 'prop-types'
import { useMutation, useQuery } from 'react-apollo'
import FormDialog from '../dialogs/FormDialog'
import ScheduleRuleForm from './ScheduleRuleForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import gql from 'graphql-tag'
import { startCase } from 'lodash-es'
import { DateTime } from 'luxon'
import { isoToClock } from './util'

const mutation = gql`
  mutation($input: ScheduleTargetInput!) {
    updateScheduleTarget(input: $input)
  }
`
const query = gql`
  query($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
    }
  }
`

export default function ScheduleRuleCreateDialog(props) {
  const { scheduleID, targetType, onClose } = props
  const [value, setValue] = useState({
    targetID: '',
    rules: [
      {
        start: DateTime.local()
          .startOf('day')
          .toUTC()
          .toISO(),
        end: DateTime.local()
          .plus({ day: 1 })
          .startOf('day')
          .toUTC()
          .toISO(),
        weekdayFilter: [true, true, true, true, true, true, true],
      },
    ],
  })

  const [mutate, mutationStatus] = useMutation(mutation, {
    onCompleted: onClose,
  })
  const { data, ...queryStatus } = useQuery(query, {
    variables: { id: scheduleID },
  })

  return (
    <FormDialog
      onClose={onClose}
      title={`Add ${startCase(targetType)} to Schedule`}
      errors={nonFieldErrors(mutationStatus.error)}
      maxWidth='md'
      onSubmit={() => {
        mutate({
          variables: {
            input: {
              target: {
                type: targetType,
                id: value.targetID,
              },
              scheduleID,

              rules: value.rules.map(r => ({
                ...r,
                start: isoToClock(r.start, data.schedule.timeZone),
                end: isoToClock(r.end, data.schedule.timeZone),
              })),
            },
          },
        })
      }}
      form={
        <ScheduleRuleForm
          targetType={targetType}
          scheduleID={scheduleID}
          disabled={queryStatus.loading || mutationStatus.loading}
          errors={fieldErrors(mutationStatus.error)}
          value={value}
          onChange={setValue}
        />
      }
    />
  )
}

ScheduleRuleCreateDialog.propTypes = {
  scheduleID: p.string.isRequired,
  targetType: p.oneOf(['rotation', 'user']).isRequired,
  onClose: p.func,
}
