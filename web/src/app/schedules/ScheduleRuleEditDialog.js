import React, { useState } from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import FormDialog from '../dialogs/FormDialog'
import ScheduleRuleForm from './ScheduleRuleForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import _ from 'lodash'
import Query from '../util/Query'
import { gqlClockTimeToISO, isoToGQLClockTime } from './util'

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
  const [value, setValue] = useState(null)

  function renderDialog(data, commit, status, zone) {
    const defaults = {
      targetID: props.target.id,
      rules: data.rules.map((r) => ({
        id: r.id,
        weekdayFilter: r.weekdayFilter,
        start: gqlClockTimeToISO(r.start, zone),
        end: gqlClockTimeToISO(r.end, zone),
      })),
    }

    function renderMutation(data, zone) {
      return (
        <Mutation mutation={mutation} onCompleted={props.onClose}>
          {(commit, status) => renderDialog(data, commit, status, zone)}
        </Mutation>
      )
    }

    return (
      <FormDialog
        onClose={props.onClose}
        title={`Edit Rules for ${_.startCase(props.target.type)}`}
        errors={nonFieldErrors(status.error)}
        maxWidth='md'
        onSubmit={() => {
          if (!value) {
            // no changes
            props.onClose()
            return
          }
          commit({
            variables: {
              input: {
                target: props.target,
                scheduleID: props.scheduleID,

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
            targetType={props.target.type}
            targetDisabled
            scheduleID={props.scheduleID}
            disabled={status.loading}
            errors={fieldErrors(status.error)}
            value={value || defaults}
            onChange={(value) => setValue({ value })}
          />
        }
      />
    )
  }

  return (
    <Query
      query={query}
      variables={{ id: props.scheduleID, tgt: props.target }}
      noPoll
      fetchPolicy='network-only'
      render={({ data }) =>
        renderMutation(data.schedule.target, data.schedule.timeZone)
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
