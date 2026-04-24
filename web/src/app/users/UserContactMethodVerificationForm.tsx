import React, { useEffect } from 'react'
import { useMutation, gql } from 'urql'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import LoadingButton from '../loading/components/LoadingButton'
import { FormContainer, FormField } from '../forms'

/*
 * Triggers sending a verification code to the specified cm
 * when the dialog is first opened
 */
const sendVerificationCodeMutation = gql`
  mutation sendContactMethodVerification(
    $input: SendContactMethodVerificationInput!
  ) {
    sendContactMethodVerification(input: $input)
  }
`

interface UserContactMethodVerificationFormProps {
  contactMethodID: string
  disabled: boolean
  errors?: Error[]
  onChange: (value: { code: string }) => void
  setSendError: (err: string) => void
  value: {
    code: string
  }
}
export default function UserContactMethodVerificationForm(
  props: UserContactMethodVerificationFormProps,
): React.ReactNode {
  const [sendCodeStatus, sendCode] = useMutation(sendVerificationCodeMutation)

  function sendAndCatch(): void {
    // Clear error on new actions.
    props.setSendError('')
    sendCode({
      input: {
        contactMethodID: props.contactMethodID,
      },
    }).catch((err) => props.setSendError(err.message))
  }

  // Attempt to send a code on load, but it's ok if it fails.
  //
  // We only want to display an error in response to a user action.
  useEffect(() => {
    sendCode({
      input: {
        contactMethodID: props.contactMethodID,
      },
    }).catch(() => {})
  }, [])

  return (
    <FormContainer optionalLabels {...props}>
      <Grid container spacing={2}>
        <Grid sx={{ display: 'flex', alignItems: 'center' }}>
          <LoadingButton
            loading={sendCodeStatus.fetching}
            disabled={props.disabled}
            buttonText='Resend Code'
            noSubmit
            onClick={() => sendAndCatch()}
          />
        </Grid>
        <Grid sx={{ flexGrow: 1 }}>
          <FormField
            fullWidth
            name='code'
            label='Verification Code'
            required
            component={TextField}
            type='number'
            step='1'
            mapOnChangeValue={(value: number) => value.toString()}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
