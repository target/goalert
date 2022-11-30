import React, { useState } from 'react'
import { gql, useMutation} from '@apollo/client'
import { useQuery } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import ScheduleRuleForm from './ScheduleRuleForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import _ from 'lodash'
import { gqlClockTimeToISO, isoToGQLClockTime } from './util'
import { ScheduleRuleInput, Target } from '../../schema'
import { GenericError } from '../error-pages'
import Query from '../util/Query'


interface ScheduleRuleEditDialogProps {
  scheduleID: string
  target: Target
  onClose: () => void
}

const query = gql`
  query ($id: ID!, $tgt: TargetInput!) {
    schedule(id: $id) {
      id
      timeZone
      target(input: $tgt) {
        rules {
          id
          start
          end
          weekdayFilter
        }
      }
    }
  }
`

const mutation = gql`
  mutation ($input: ScheduleTargetInput!) {
    updateScheduleTarget(input: $input)
  }
`

export default function ScheduleRuleEditDialog(
  props: ScheduleRuleEditDialogProps,
): JSX.Element {
  const onClose = props.onClose
  const target = props.target
  const scheduleID = props.scheduleID
  const [value, setValue] = useState(null)
  const [{ data, error, fetching }] = useQuery({
    query,
    variables: { id: props.scheduleID, tgt: props.target },
  })
  const zone = data.schedule.TimeZone
  const [deleteRule] = useMutation(mutation, {
    onCompleted: props.onClose,
  })
  if (error) {
    return <GenericError error={error.message} />
  }
  const defaults = {
    targetID: target.id,
    rules: data.rules.map((r: ScheduleRuleInput) => ({
      id: r.id,
      weekdayFilter: r.weekdayFilter,
      start: gqlClockTimeToISO(r.start, zone),
      end: gqlClockTimeToISO(r.end, zone),
    })),
  }
    return (
      <FormDialog
        onClose={onClose}
        title={`Edit Rules for ${_.startCase(target.type)}`}
        errors={nonFieldErrors(error)}
        maxWidth='md'
        onSubmit={() => {
          if (!value) {
            // no changes
            onClose()
            return
          }
          commit({
            variables: {
              input: {
                target,
                scheduleID,

                rules: value.rules.map((r) => ({
                  ...r,
                  start: isoToGQLClockTime(r.start, zone),
                  end: isoToGQLClockTime(r.end, zone),
                })),
              },
            },
          })
        }}
        form={
          <ScheduleRuleForm
            targetType={target.type}
            targetDisabled
            scheduleID={scheduleID}
            disabled={loading}
            errors={fieldErrors(error)}
            value={value || defaults}
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }

  return (
    <Query
      query={query}
      variables={{ id: scheduleID, tgt: target }}
      noPoll
      fetchPolicy='network-only'
      render={({ data }) =>
        renderMutation(data.schedule.target, data.schedule.timeZone)
      }
    />
  )
}

// ScheduleRuleEditDialog.propTypes = {
//   scheduleID: p.string.isRequired,
//   target: p.shape({
//     type: p.oneOf(['rotation', 'user']).isRequired,
//     id: p.string.isRequired,
//   }).isRequired,
//   onClose: p.func,
// }
