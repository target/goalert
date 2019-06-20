import React, { useState } from 'react'
import p from 'prop-types'
import FormControl from '@material-ui/core/FormControl'
import FormHelperText from '@material-ui/core/FormHelperText'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import LoadingButton from '../loading/components/LoadingButton'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { makeStyles } from '@material-ui/core/styles'

/*
 * Triggers sending a verification code to the specified cm
 */
const sendContactMethodVerificationMutation = gql`
  mutation SendContactMethodVerificationMutation(
    $input: SendContactMethodVerificationInput
  ) {
    sendContactMethodVerification(input: $input) {
      id
    }
  }
`

const useStyles = makeStyles({
  fieldGridItem: {
    flexGrow: 1,
  },
  fieldStyle: {
    width: '100%',
  },
  sendGridItem: {
    display: 'flex',
    alignItems: 'center',
  },
})

export default function ContactMethodVerificationForm(props) {
  const [sendAttempted, setSendAttempted] = useState(false)
  const classes = useStyles()

  function getTitle() {
    if (sendAttempted) {
      return 'Resend Code'
    } else {
      return 'Send Code'
    }
  }

  return (
    <Grid container spacing={2}>
      <Grid item className={classes.sendGridItem}>
        <Mutation
          mutation={sendContactMethodVerificationMutation}
          onError={() =>
            props.setSendError('Too many messages! Try again after some time.')
          }
        >
          {(commit, status) => (
            <LoadingButton
              color='primary'
              loading={status.loading}
              disabled={props.disabled}
              buttonText={getTitle()}
              onClick={() => {
                setSendAttempted(true)

                commit({
                  variables: {
                    input: {
                      contact_method_id: props.contactMethodID,
                    },
                  },
                })
              }}
            />
          )}
        </Mutation>
      </Grid>
      <Grid item className={classes.fieldGridItem}>
        <FormControl
          className={classes.fieldStyle}
          disabled={props.disabled}
          error={!!props.error.length}
        >
          <TextField
            aria-label='Code'
            disabled={props.disabled}
            error={!!props.error.length}
            label='Verification Code'
            name='code'
            onChange={e => props.onChange(e.target.value)}
            placeholder='Enter the verification code received'
            value={props.value}
          />
          <FormHelperText>{props.error[0]}</FormHelperText>
        </FormControl>
      </Grid>
    </Grid>
  )
}

ContactMethodVerificationForm.propTypes = {
  contactMethodID: p.string.isRequired,
  disabled: p.bool.isRequired,
  error: p.array.isRequired,
  onChange: p.func.isRequired,
  setSendError: p.func.isRequired,
  value: p.string.isRequired,
}
