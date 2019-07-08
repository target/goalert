import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import LoadingButton from '../loading/components/LoadingButton'
import { Mutation } from 'react-apollo'
import { makeStyles } from '@material-ui/core/styles'
import { graphql2Client } from '../apollo'
import { FormContainer, FormField } from '../forms'
import { sendVerificationCodeMutation } from './UserContactMethodVerificationDialog'

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

  function getTitle() {
    if (props.sendAttempted) {
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
            mutation={sendVerificationCodeMutation}
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
                  props.setSendAttempted(true) // changes text of send button to "resend"

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
            type='number'
            step='1'
            mapOnChangeValue={value => value.toString()}
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
  sendAttempted: p.bool.isRequired,
  setSendAttempted: p.func.isRequired,
  value: p.shape({
    code: p.string.isRequired,
  }).isRequired,
}
