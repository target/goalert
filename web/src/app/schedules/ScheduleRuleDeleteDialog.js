import React from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors } from '../util/errutil'
import { startCase } from 'lodash'
import Query from '../util/Query'

const query = gql`
  query ($id: ID!, $tgt: TargetInput!) {
    schedule(id: $id) {
      id
      target(input: $tgt) {
        target {
          id
          name
          type
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

export default function ScheduleRuleDeleteDialog(props) {
  function renderDialog(data, commit, status) {
    return (
      <FormDialog
        onClose={props.onClose}
        title={`Remove ${startCase(props.target.type)} From Schedule?`}
        subTitle={`This will remove all rules, as well as end any active or future on-call shifts on this schedule for ${props.target.type}: ${data.target.name}.`}
        caption='Overrides will not be affected.'
        confirm
        errors={nonFieldErrors(status.error)}
        onSubmit={() => {
          commit({
            variables: {
              input: {
                target: props.target,
                scheduleID: props.scheduleID,

                rules: [],
              },
            },
          })
        }}
      />
    )
  }

  function renderMutation(data) {
    return (
      <Mutation mutation={mutation} onCompleted={props.onClose}>
        {(commit, status) => renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  return (
    <Query
      query={query}
      variables={{
        id: props.scheduleID,
        tgt: props.target,
      }}
      noPoll
      render={({ data }) => renderMutation(data.schedule.target)}
    />
  )
}

ScheduleRuleDeleteDialog.propTypes = {
  scheduleID: p.string.isRequired,
  target: p.shape({ id: p.string.isRequired, type: p.string.isRequired })
    .isRequired,
  onClose: p.func,
}
