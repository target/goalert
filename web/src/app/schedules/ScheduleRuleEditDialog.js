import React, { useState } from 'react'
import { useQuery, gql, useMutation } from 'urql'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import ScheduleRuleForm from './ScheduleRuleForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import _ from 'lodash'
import { gqlClockTimeToISO, isoToGQLClockTime } from './util'
import { useScheduleTZ } from './useScheduleTZ'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

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

export default function ScheduleRuleEditDialog(props) {
  const [state, setState] = useState({ value: null })

  const [{ data, fetching, error: readError }] = useQuery({
    query,
    requestPolicy: 'network-only',
    variables: {
      id: props.scheduleID,
      tgt: props.target,
    },
  })
  const [{ error }, commit] = useMutation(mutation)
  const { zone } = useScheduleTZ(props.scheduleID)

  if (readError) return <GenericError error={readError.message} />
  if (fetching && !data) return <Spinner />

  const defaults = {
    targetID: props.target.id,
    rules: data.schedule.target.rules.map((r) => ({
      id: r.id,
      weekdayFilter: r.weekdayFilter,
      start: gqlClockTimeToISO(r.start, zone),
      end: gqlClockTimeToISO(r.end, zone),
    })),
  }
  return (
    <FormDialog
      onClose={props.onClose}
      title={`Edit Rules for ${_.startCase(props.target.type)}`}
      errors={nonFieldErrors(error)}
      maxWidth='md'
      onSubmit={() => {
        if (!state.value) {
          // no changes
          props.onClose()
          return
        }
        commit({
          input: {
            target: props.target,
            scheduleID: props.scheduleID,

            rules: state.value.rules.map((r) => ({
              ...r,
              start: isoToGQLClockTime(r.start, zone),
              end: isoToGQLClockTime(r.end, zone),
            })),
          },
        }).then(() => {
          props.onClose()
        })
      }}
      form={
        <ScheduleRuleForm
          targetType={props.target.type}
          targetDisabled
          scheduleID={props.scheduleID}
          disabled={fetching}
          errors={fieldErrors(error)}
          value={state.value || defaults}
          onChange={(value) => setState({ value })}
        />
      }
    />
  )
}
ScheduleRuleEditDialog.propTypes = {
  scheduleID: p.string.isRequired,
  target: p.shape({
    type: p.oneOf(['rotation', 'user']).isRequired,
    id: p.string.isRequired,
  }).isRequired,
  onClose: p.func,
}
