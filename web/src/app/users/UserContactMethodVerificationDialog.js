import React, { useEffect, useState } from 'react'
import { useMutation } from 'react-apollo-hooks'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import Query from '../util/Query'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import UserContactMethodVerificationForm from './UserContactMethodVerificationForm'
import { graphql2Client } from '../apollo'
import { formatPhoneNumber } from './util'
import { Config } from '../util/RequireConfig'

/*
 * Triggers sending a verification code to the specified cm
 * when the dialog is first opened
 */
export const sendVerificationCodeMutation = gql`
  mutation sendContactMethodVerification(
    $input: SendContactMethodVerificationInput!
  ) {
    sendContactMethodVerification(input: $input)
  }
`

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

export default function UserContactMethodVerificationDialog(props) {
  // state initialization
  const [value, setValue] = useState({
    code: '',
  })
  const [sendError, setSendError] = useState('')

  const sendCode = useMutation(sendVerificationCodeMutation, {
    // mutation options
    variables: {
      input: {
        contactMethodID: props.contactMethodID,
      },
    },
  })

  // componentDidMount
  useEffect(() => {
    sendCode().catch(err => setSendError(err.message))
  }, [])

  // dialog rendered that handles rendering the verification form
  function renderDialog(commit, status, cm) {
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    return (
      <Config>
        {config => {
          const fromNumber = config['Twilio.FromNumber']

          let caption = null
          if (fromNumber && cm.type === 'SMS') {
            caption = `If you do not receive a code, try sending UNSTOP to ${formatPhoneNumber(
              fromNumber,
            )} before resending.`
          }

          return (
            <FormDialog
              title='Verify Contact Method'
              subTitle={`A verification code has been sent to ${formatPhoneNumber(
                cm.value,
              )} (${cm.type})`}
              caption={caption}
              loading={loading}
              errors={
                sendError ? [{ message: sendError }] : nonFieldErrors(error)
              }
              onClose={props.onClose}
              onSubmit={() =>
                commit({
                  variables: {
                    input: {
                      contactMethodID: cm.id,
                      code: value.code,
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
        }}
      </Config>
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
