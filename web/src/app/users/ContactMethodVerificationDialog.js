import React, { useState } from 'react'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import Query from '../util/Query'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import ContactMethodVerificationForm from './ContactMethodVerificationForm'

/*
 * Reactivates a cm if disabled and the verification code matches
 */
const verifyContactMethodMutation = gql`
  mutation VerifyContactMethodMutation($input: VerifyContactMethodInput) {
    verifyContactMethod(input: $input) {
      contact_method_ids
    }
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

export default function ContactMethodVerificationDialog(props) {
  const [value, setValue] = useState('')
  const [sendError, setSendError] = useState('')

  // dialog rendered that handles rendering the verification form
  function renderDialog(commit, status, cm) {
    const { loading, error } = status
    const fieldErr = fieldErrors(error)

    return (
      <FormDialog
        title={`Verify Contact Method by ${cm.type}`}
        subtitle={`Verifying "${cm.name}" at ${formatNumber(cm.value)}`}
        loading={loading}
        errors={sendError || nonFieldErrors(error)}
        onClose={props.onClose}
        onSubmit={() =>
          commit({
            variables: {
              input: {
                contact_method_id: cm.id,
                verification_code: parseInt(value),
              },
            },
          })
        }
        form={
          <ContactMethodVerificationForm
            contactMethodID={cm.id}
            error={fieldErr}
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
        mutation={verifyContactMethodMutation}
        // todo: awaitRefetchQueries
        // todo: refetchQueries={['cmList']}
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

ContactMethodVerificationDialog.propTypes = {
  onClose: p.func.isRequired,
  contactMethodID: p.string.isRequired,
}
