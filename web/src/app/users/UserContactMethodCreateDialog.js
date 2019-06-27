import React from 'react'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import ContactMethodVerificationForm from './ContactMethodVerificationForm'

const createMutation = gql`
  mutation($input: CreateUserContactMethodInput!) {
    createUserContactMethod(input: $input) {
      id
    }
  }
`
const verifyContactMethodMutation = gql`
  mutation verifyContactMethod($input: VerifyContactMethodInput!) {
    verifyContactMethod(input: $input)
  }
`

export default class UserContactMethodCreateDialog extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    // values for contact method form
    cmValue: {
      name: '',
      type: 'SMS',
      value: '+1',
    },
    // value for verification form
    verValue: {
      code: '',
    },
    sendError: '', // error if verification code send fails
    errors: [],
    contactMethodID: null, // used for verification mutation
  }

  onComplete = data => {
    if (this.state.contactMethodID) {
      this.props.onClose()
    } else {
      this.setState({
        contactMethodID: data.createUserContactMethod.id, // output from create mutation
      })
    }
  }

  getInputVariables() {
    if (this.state.contactMethodID) {
      return {
        variables: {
          input: {
            contactMethodID: this.state.contactMethodID,
            verificationCode: this.state.verValue.code,
          },
        },
      }
    } else {
      return {
        variables: {
          input: {
            ...this.state.cmValue,
            userID: this.props.userID,
            newUserNotificationRule: {
              delayMinutes: 0,
            },
          },
        },
      }
    }
  }

  render() {
    return (
      <Mutation
        client={graphql2Client}
        mutation={
          this.state.contactMethodID
            ? verifyContactMethodMutation
            : createMutation
        }
        awaitRefetchQueries
        refetchQueries={['nrList', 'cmList']}
        onCompleted={this.onComplete}
      >
        {(commit, status) => this.renderDialog(commit, status)}
      </Mutation>
    )
  }

  renderDialog(commit, status) {
    const { loading, error } = status
    const { contactMethodID } = this.state

    return (
      <FormDialog
        title='Create New Contact Method'
        subTitle={contactMethodID ? 'Verify contact method to continue' : null}
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          const input = this.getInputVariables()
          return commit(input)
        }}
        form={this.renderForm(status)}
      />
    )
  }

  renderForm(status) {
    // these values are different depending on which
    // mutation wraps the dialog
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    if (this.state.contactMethodID) {
      return (
        <ContactMethodVerificationForm
          contactMethodID={this.state.contactMethodID}
          disabled={loading}
          errors={fieldErrs}
          onChange={verValue => this.setState({ verValue })}
          setSendError={sendError => this.setState({ sendError })}
          value={this.state.verValue}
        />
      )
    } else {
      return (
        <UserContactMethodForm
          disabled={loading}
          errors={fieldErrs}
          onChange={cmValue => this.setState({ cmValue })}
          value={this.state.cmValue}
        />
      )
    }
  }
}
