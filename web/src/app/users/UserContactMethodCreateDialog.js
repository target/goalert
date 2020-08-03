import React, { useState } from 'react'
import p from 'prop-types'
import { AppLink } from '../util/AppLink'
import gql from 'graphql-tag'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import { useMutation, useQuery } from '@apollo/react-hooks'

const query = gql`
  query userName($id: ID!) {
    user(id: $id) {
      name
    }
  }
`
const createMutation = gql`
  mutation($input: CreateUserContactMethodInput!) {
    createUserContactMethod(input: $input) {
      id
    }
  }
`

export default function UserContactMethodCreateDialog(props) {
  // values for contact method form
  const [conflictingUserID, setConflictingUserID] = useState('')
  const [cmValue, setCmValue] = useState({
    name: '',
    type: 'SMS',
    value: '',
  })

  const [createCM, createCMStatus] = useMutation(createMutation, {
    onCompleted: (result) => {
      props.onClose({ contactMethodID: result.createUserContactMethod.id })
    },
    onError: (err) => {
      const errorMessage = err.message || err
      const message = errorMessage.split(' ')
      const userID = message.pop()
      if (userID.length > 15) {
        setConflictingUserID(userID)
      }
    },
    variables: {
      input: {
        ...cmValue,
        userID: props.userID,
        newUserNotificationRule: {
          delayMinutes: 0,
        },
      },
    },
  })
  const { data } = useQuery(query, {
    variables: {
      id: conflictingUserID,
    },
    skip: Boolean(!conflictingUserID), // skip query if no conflicting user
  })
  const { loading, error } = createCMStatus
  let messageErr = `Contact method already exists for that type and value ${
    data?.user?.name ? 'by user ' + data.user.name : ''
  }`
  if (conflictingUserID)
    messageErr = (
      <AppLink to={`/users/${conflictingUserID}`}>{messageErr}</AppLink>
    )
  const conflictingUserErrorMessage = {
    field: 'value',
    message: messageErr,
    details: '',
    path: 'createUserContactMethod',
  }
  const fieldErrs = conflictingUserID
    ? [conflictingUserErrorMessage]
    : fieldErrors(error)
  const { title = 'Create New Contact Method', subtitle } = props

  const form = (
    <UserContactMethodForm
      disabled={loading}
      errors={fieldErrs}
      onChange={(cmValue) => setCmValue(cmValue)}
      value={cmValue}
      disclaimer={props.disclaimer}
    />
  )
  return (
    <FormDialog
      data-cy='create-form'
      title={title}
      subTitle={subtitle}
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      // wrapped to prevent event from passing into createCM
      onSubmit={() => createCM()}
      form={form}
    />
  )
}

UserContactMethodCreateDialog.propTypes = {
  userID: p.string.isRequired,
  onClose: p.func,
  disclaimer: p.string,
  title: p.string,
  subtitle: p.string,
}
