import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import FormDialog from '../dialogs/FormDialog'
import { Mutation } from 'react-apollo'
import { graphql2Client } from '../apollo'
import { nonFieldErrors } from '../util/errutil'

const mutation = gql`
  mutation($id: ID!) {
    deleteAll(input: [{ id: $id, type: contactMethod }])
  }
`
export default class UserContactMethodDeleteDialog extends React.PureComponent {
  static propTypes = {
    contactMethodID: p.string.isRequired,
    onClose: p.func.isRequired, // passed to FormDialog
  }

  render() {
    return (
      <Mutation
        mutation={mutation}
        client={graphql2Client}
        onCompleted={this.props.onClose}
        awaitRefetchQueries
        refetchQueries={['nrList', 'cmList']}
      >
        {(commit, status) => this.renderDialog(commit, status)}
      </Mutation>
    )
  }

  renderDialog(commit, { loading, error }) {
    const { contactMethodID, ...rest } = this.props
    return (
      <FormDialog
        title='Are you sure?'
        confirm
        loading={loading}
        errors={nonFieldErrors(error)}
        subTitle='This will delete the contact method.'
        caption='This will also delete any notification rules associated with this contact method.'
        onSubmit={() => commit({ variables: { id: contactMethodID } })}
        {...rest}
      />
    )
  }
}
