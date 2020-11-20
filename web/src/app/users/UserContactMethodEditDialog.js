import React from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import Query from '../util/Query'
import { pick } from 'lodash'

const query = gql`
  query($id: ID!) {
    userContactMethod(id: $id) {
      id
      name
      type
      value
    }
  }
`

const mutation = gql`
  mutation($input: UpdateUserContactMethodInput!) {
    updateUserContactMethod(input: $input)
  }
`

export default class UserContactMethodEditDialog extends React.PureComponent {
  static propTypes = {
    contactMethodID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: null,
    errors: [],
    edit: true,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.contactMethodID }}
        render={({ data }) => this.renderMutation(data.userContactMethod)}
        noPoll
      />
    )
  }

  renderMutation({ name, type, value }) {
    return (
      <Mutation mutation={mutation} onCompleted={this.props.onClose}>
        {(commit, status) =>
          this.renderDialog(commit, status, { name, type, value })
        }
      </Mutation>
    )
  }

  renderDialog(commit, status, defaultValue) {
    const { loading, error } = status

    const fieldErrs = fieldErrors(error)

    return (
      <FormDialog
        title='Edit Contact Method'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              // only pass 'name'
              input: {
                ...pick(this.state.value, 'name'),
                id: this.props.contactMethodID,
              },
            },
          })
        }}
        form={
          <UserContactMethodForm
            errors={fieldErrs}
            disabled={loading}
            edit={this.state.edit}
            value={this.state.value || defaultValue}
            onChange={(value) => this.setState({ value })}
          />
        }
      />
    )
  }
}
