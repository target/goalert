import React, { useEffect } from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import LoadingButton from '../loading/components/LoadingButton'
import gql from 'graphql-tag'
import { makeStyles } from '@material-ui/core/styles'
import { FormContainer, FormField } from '../forms'
import { useMutation } from '@apollo/react-hooks'

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

const useStyles = makeStyles({
  fieldGridItem: {
    flexGrow: 1,
  },
  sendGridItem: {
    display: 'flex',
    alignItems: 'center',
  },
})

export default function UserContactMethodVerificationForm(props) {
  const classes = useStyles()

  const [sendCode, sendCodeStatus] = useMutation(sendVerificationCodeMutation, {
    variables: {
      input: {
        contactMethodID: props.contactMethodID,
      },
    },
  })

  function sendAndCatch() {
    // Clear error on new actions.
    props.setSendError(null)
    sendCode().catch((err) => props.setSendError(err.message))
  }

  // Attempt to send a code on load, but it's ok if it fails.
  //
  // We only want to display an error in response to a user action.
  useEffect(() => {
    sendCode().catch(() => {})
  }, [])

  return (
    <FormContainer optionalLabels {...props}>
      <Grid container spacing={2}>
        <Grid item className={classes.sendGridItem}>
          <LoadingButton
            color='primary'
            loading={sendCodeStatus.loading}
            disabled={props.disabled}
            buttonText='Resend Code'
            noSubmit
            onClick={() => sendAndCatch()}
          />
        </Grid>
        <Grid item className={classes.fieldGridItem}>
          <FormField
            fullWidth
            name='code'
            label='Verification Code'
            required
            component={TextField}
            type='number'
            step='1'
            mapOnChangeValue={(value) => value.toString()}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}

UserContactMethodVerificationForm.propTypes = {
  contactMethodID: p.string.isRequired,
  disabled: p.bool.isRequired,
  errors: p.arrayOf(
    p.shape({
      field: p.oneOf(['code']).isRequired,
      message: p.string.isRequired,
    }),
  ),
  onChange: p.func.isRequired,
  setSendError: p.func.isRequired,
  value: p.shape({
    code: p.string.isRequired,
  }).isRequired,
}
