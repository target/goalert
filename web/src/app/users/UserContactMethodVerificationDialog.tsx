import React, { useState } from 'react'
import { useMutation, useQuery, gql } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import UserContactMethodVerificationForm from './UserContactMethodVerificationForm'

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
  query ($id: ID!) {
    userContactMethod(id: $id) {
      id
      type
      formattedValue
      lastVerifyMessageState {
        status
        details
        formattedSrcValue
      }
    }
  }
`

interface UserContactMethodVerificationDialogProps {
  onClose: () => void
  contactMethodID: string
}

const noSuspense = { suspense: false }
export default function UserContactMethodVerificationDialog(
  props: UserContactMethodVerificationDialogProps,
): React.ReactNode {
  const [value, setValue] = useState<{ code: string }>({
    code: '',
  })
  const [sendError, setSendError] = useState('')

  const [status, submitVerify] = useMutation(verifyContactMethodMutation)

  const [{ data }] = useQuery({
    query: contactMethodQuery,
    variables: { id: props.contactMethodID },
    context: noSuspense,
  })

  const fromNumber =
    data?.userContactMethod?.lastVerifyMessageState?.formattedSrcValue ?? ''
  const cm = data?.userContactMethod ?? {}

  const { fetching, error } = status
  const fieldErrs = fieldErrors(error)

  let caption = null
  if (fromNumber && cm.type === 'SMS') {
    caption = `If you do not receive a code, try sending START to ${fromNumber} before resending.`
  }
  return (
    <FormDialog
      title='Verify Contact Method'
      subTitle={`A verification code has been sent to ${cm.formattedValue} (${cm.type})`}
      caption={caption}
      loading={fetching || !cm.type}
      errors={
        sendError
          ? [new Error(sendError)].concat(nonFieldErrors(error))
          : nonFieldErrors(error)
      }
      data-cy='verify-form'
      onClose={props.onClose}
      onSubmit={() => {
        setSendError('')
        submitVerify(
          {
            input: {
              contactMethodID: props.contactMethodID,
              code: value.code,
            },
          },
          { additionalTypenames: ['UserContactMethod'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }}
      form={
        <UserContactMethodVerificationForm
          contactMethodID={props.contactMethodID}
          errors={fieldErrs}
          setSendError={setSendError}
          disabled={fetching}
          value={value}
          onChange={(value: { code: string }) => setValue(value)}
        />
      }
    />
  )
}
