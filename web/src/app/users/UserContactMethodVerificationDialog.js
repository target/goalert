import React, { useState } from 'react'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import Query from '../util/Query'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import UserContactMethodVerificationForm from './UserContactMethodVerificationForm'
import { graphql2Client } from '../apollo'

/*
 * Reactivates a cm if disabled and the verification code matches
 */
const verifyContactMethodMutation = gql`
  mutation verifyContactMethod($input: VerifyContactMethodInput!) {
    verifyContactMethod(input: $input)
  }
`

/*
 * Get cm data so this component isn't dependent on parent props
 */
const contactMethodQuery = gql`
  query($id: ID!) {
    userContactMethod(id: $id) {
      id
      name
      type
      value
    }
  }
`

/*
 * Used for the subtitle of the dialog
 */
function formatNumber(n) {
  if (n.startsWith('+1')) {
    return `+1 (${n.slice(2, 5)}) ${n.slice(5, 8)}-${n.slice(8)}`
  }
  if (n.startsWith('+91')) {
    return `+91-${n.slice(3, 5)}-${n.slice(5, 8)}-${n.slice(8)}`
  }
  if (n.startsWith('+44')) {
    return `+44 ${n.slice(3, 7)} ${n.slice(7)}`
  } else {
    return <span>{n}</span>
  }
}

export default function UserContactMethodVerificationDialog(props) {
  const [value, setValue] = useState({
    code: '',
  })
  const [sendError, setSendError] = useState('')

  // dialog rendered that handles rendering the verification form
  function renderDialog(commit, status, cm) {
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    return (
      <FormDialog
        title={`Verify Contact Method by ${cm.type}`}
        subtitle={`Verifying "${cm.name}" at ${formatNumber(cm.value)}`}
        loading={loading}
        errors={nonFieldErrors(error) || [{ message: sendError }]}
        onClose={props.onClose}
        onSubmit={() =>
          commit({
            variables: {
              input: {
                contactMethodID: cm.id,
                verificationCode: value.code,
              },
            },
          })
        }
        form={
          <UserContactMethodVerificationForm
            contactMethodID={cm.id}
            errors={fieldErrs}
            setSendError={setSendError}
            disabled={loading}
            value={value}
            onChange={value => setValue(value)}
          />
        }
      />
    )
  }

  // wraps the dialog with the mutation
  function renderMutation(cm) {
    return (
      <Mutation
        client={graphql2Client}
        mutation={verifyContactMethodMutation}
        awaitRefetchQueries
        refetchQueries={['cmList']}
        onCompleted={props.onClose}
      >
        {(commit, status) => renderDialog(commit, status, cm)}
      </Mutation>
    )
  }

  // queries for cm data for the dialog subtitle
  return (
    <Query
      query={contactMethodQuery}
      variables={{ id: props.contactMethodID }}
      render={({ data }) => renderMutation(data.userContactMethod)}
      noPoll
    />
  )
}

UserContactMethodVerificationDialog.propTypes = {
  onClose: p.func.isRequired,
  contactMethodID: p.string.isRequired,
}
