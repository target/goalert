import React, { useState } from 'react'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import UserContactMethodVerificationForm from './UserContactMethodVerificationForm'
import { formatPhoneNumber } from './util'

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

  const onComplete = (data, errors) => {
    if (!contactMethodID) {
      setContactMethodID(data.createUserContactMethod.id)
    } else {
      props.onClose()
    }
  }

  function getInputVariables() {
    if (contactMethodID) {
      return {
        variables: {
          input: {
            contactMethodID: contactMethodID,
            code: verValue.code,
          },
        },
      }
    } else {
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
  }

  function renderDialog(commit, status) {
    const { loading, error } = status
    return (
      <FormDialog
        data-cy={contactMethodID ? 'verify-form' : 'create-form'}
        title='Create New Contact Method'
        subTitle={
          contactMethodID
            ? `A verification code has been sent to ${formatPhoneNumber(
                cmValue.value,
              )} (${cmValue.type})`
            : null
        }
        loading={loading}
        errors={sendError ? [{ message: sendError }] : nonFieldErrors(error)}
        onClose={props.onClose}
        onSubmit={() => commit(getInputVariables())}
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
          onChange={verValue => setVerValue(verValue)}
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
      onCompleted={onComplete}
    >
      {(commit, status) => renderDialog(commit, status)}
    </Mutation>
  )
}

UserContactMethodCreateDialog.propTypes = {
  userID: p.string.isRequired,
  onClose: p.func,
}
