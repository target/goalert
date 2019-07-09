import React, { useState } from 'react'
import { useMutation } from 'react-apollo-hooks'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import UserContactMethodVerificationForm from './UserContactMethodVerificationForm'
import { sendVerificationCodeMutation } from './UserContactMethodVerificationDialog'

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
export default function UserContactMethodCreateDialog(props) {
  // values for contact method form
  const [cmValue, setCmValue] = useState({
    name: '',
    type: 'SMS',
    value: '+1',
  })
  // value for verification form
  const [verValue, setVerValue] = useState({ code: '' })
  const [sendError, setSendError] = useState('') // error if verification code send fails
  const [contactMethodID, setContactMethodID] = useState(null) // used for verification mutation

  function getInputVariables() {
    return {
      variables: {
        input: {
          ...cmValue,
          userID: props.userID,
          newUserNotificationRule: {
            delayMinutes: 0,
          },
        },
      },
    }
  }
  const sendCode = useMutation(sendVerificationCodeMutation, {
    // mutation options
    variables: {
      input: {
        contactMethodID: contactMethodID,
      },
    },
  })

  function renderDialog(commit, status) {
    const { loading, error } = status
    return (
      <FormDialog
        title='Create New Contact Method'
        subTitle={contactMethodID ? 'Verify contact method to continue' : null}
        loading={loading}
        errors={sendError ? [{ message: sendError }] : nonFieldErrors(error)}
        onClose={props.onClose}
        onSubmit={() => {
          const input = getInputVariables()
          return commit(input).then(() =>
            sendCode().catch(err => setSendError(err.message)),
          )
        }}
        form={renderForm(status)}
      />
    )
  }

  function renderForm(status) {
    // these values are different depending on which
    // mutation wraps the dialog
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    if (contactMethodID) {
      return (
        <UserContactMethodVerificationForm
          contactMethodID={contactMethodID}
          disabled={loading}
          errors={fieldErrs}
          onChange={verValue => setVerValue({ verValue })}
          setSendError={setSendError}
          value={verValue}
        />
      )
    } else {
      return (
        <UserContactMethodForm
          disabled={loading}
          errors={fieldErrs}
          onChange={cmValue => setCmValue(cmValue)}
          value={cmValue}
        />
      )
    }
  }

  return (
    <Mutation
      client={graphql2Client}
      mutation={contactMethodID ? verifyContactMethodMutation : createMutation}
      awaitRefetchQueries
      refetchQueries={['nrList', 'cmList']}
      onCompleted={data => setContactMethodID(data.createUserContactMethod.id)}
    >
      {(commit, status) => renderDialog(commit, status)}
    </Mutation>
  )
}

UserContactMethodCreateDialog.propTypes = {
  userID: p.string.isRequired,
  onClose: p.func,
}

/*
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
            code: this.state.verValue.code,
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
        <UserContactMethodVerificationForm
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
*/
