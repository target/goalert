import React from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import ScheduleForm from './ScheduleForm'
import { Mutation } from '@apollo/client/react/components'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import Query from '../util/Query'

const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      name
      description
      timeZone
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateScheduleInput!) {
    updateSchedule(input: $input)
  }
`

export default function ScheduleEditDialog(props) {
  ScheduleEditDialog.propTypes = {
    scheduleID: p.string.isRequired,
    onClose: p.func,
  }

  const [state, setState] = useState(null)

  const renderMutation = (data) => {
    return (
      <Mutation mutation={mutation} onCompleted={this.props.onClose}>
        {(...args) => renderForm(data, ...args)}
      </Mutation>
    )
  }

  const renderForm = (data, commit, status) => {
    return (
      <FormDialog
        onClose={props.onClose}
        title='Edit Schedule'
        errors={nonFieldErrors(status.error)}
        onSubmit={() =>
          commit({
            variables: {
              input: {
                id: props.scheduleID,
                ...state,
              },
            },
          })
        }
        form={
          <ScheduleForm
            disabled={status.loading}
            errors={fieldErrors(status.error)}
            value={
              state || {
                name: data.name,
                description: data.description,
                timeZone: data.timeZone,
              }
            }
            onChange={(value) => setState({ value })}
          />
        }
      />
    )
  }

  return (
    <Query
      query={query}
      variables={{ id: props.scheduleID }}
      render={({ data }) => renderMutation(data.schedule)}
    />
  )
}
