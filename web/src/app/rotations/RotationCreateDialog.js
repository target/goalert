import React from 'react'
import { Redirect } from 'react-router'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import RotationForm from './RotationForm'
import { DateTime } from 'luxon'

const mutation = gql`
  mutation($input: CreateRotationInput!) {
    createRotation(input: $input) {
      id
      name
      description
      start
      timeZone
      type
      shiftLength
    }
  }
`

export default class RotationCreateDialog extends React.PureComponent {
  state = {
    value: {
      name: '',
      description: '',
      timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      type: 'daily',
      start: DateTime.local()
        .startOf('hour')
        .toISO(),
      shiftLength: 1,
      favorite: true,
    },
  }

  render() {
    return (
      <Mutation mutation={mutation}>
        {(commit, status) => this.renderDialog(commit, status)}
      </Mutation>
    )
  }

  renderDialog(commit, status) {
    const { loading } = status
    if (status.data && status.data.createRotation) {
      return (
        <Redirect push to={`/rotations/${status.data.createRotation.id}`} />
      )
    }

    return (
      <FormDialog
        title='Create Rotation'
        loading={loading}
        errors={nonFieldErrors(status.error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                timeZone: this.state.value.timeZone,
                ...this.state.value,
              },
            },
          })
        }}
        form={
          <RotationForm
            errors={fieldErrors(status.error)}
            disabled={status.loading}
            value={this.state.value}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }
}
