import React, { useState } from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import LoadingButton from '../loading/components/LoadingButton'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { makeStyles } from '@material-ui/core/styles'
import { graphql2Client } from '../apollo'
import { FormContainer, FormField } from '../forms'

/*
 * Triggers sending a verification code to the specified cm
 */
const sendContactMethodVerificationMutation = gql`
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
    <FormContainer optionalLabels {...props}>
      <Grid container spacing={2}>
        <Grid item className={classes.sendGridItem}>
          <Mutation
            client={graphql2Client}
            mutation={sendContactMethodVerificationMutation}
            onError={() =>
              props.setSendError(
                'Too many messages! Try again after some time.',
              )
            }
          >
            {(commit, status) => (
              <LoadingButton
                color='primary'
                loading={status.loading}
                disabled={props.disabled}
                buttonText={getTitle()}
                onClick={() => {
                  setSendAttempted(true) // changes text of send button to "resend"

                  commit({
                    variables: {
                      input: {
                        contactMethodID: props.contactMethodID,
                      },
                    },
                  })
                }}
              />
            )}
          </Mutation>
        </Grid>
        <Grid item className={classes.fieldGridItem}>
          <FormField
            fullWidth
            name='code'
            label='Verification Code'
            required
            component={TextField}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}

ContactMethodVerificationForm.propTypes = {
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
