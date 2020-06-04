import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'
import FormDialog from '../dialogs/FormDialog'
import RotationForm from './RotationForm'

const query = gql`
  query($id: ID!) {
    rotation(id: $id) {
      id
      name
      description
      timeZone
      type
      shiftLength
      start
    }
  }
`

const mutation = gql`
  mutation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

export default class RotationEditDialog extends React.PureComponent {
  static propTypes = {
    rotationID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: null,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.rotationID }}
        noPoll
        render={({ data }) => this.renderMutation(data.rotation)}
      />
    )
  }

  renderMutation(data) {
    return (
      <Mutation mutation={mutation} onCompleted={this.props.onClose}>
        {(...args) => this.renderForm(data, ...args)}
      </Mutation>
    )
  }

  renderForm = (data, commit, status) => {
    return (
      <FormDialog
        title='Edit Rotation'
        errors={nonFieldErrors(status.error)}
        onClose={this.props.onClose}
        onSubmit={() =>
          commit({
            variables: {
              input: {
                id: this.props.rotationID,
                ...this.state.value,
              },
            },
          })
        }
        form={
          <RotationForm
            errors={fieldErrors(status.error)}
            disabled={status.loading}
            value={
              this.state.value || {
                name: data.name,
                description: data.description,
                timeZone: data.timeZone,
                type: data.type,
                shiftLength: data.shiftLength,
                start: data.start,
              }
            }
            onChange={(value) => this.setState({ value })}
          />
        }
      />
    )
  }
}
