import React, { Component } from 'react'
import p from 'prop-types'
import FormControl from '@material-ui/core/FormControl'
import FormHelperText from '@material-ui/core/FormHelperText'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import LoadingButton from '../../loading/components/LoadingButton'
import ApolloFormDialog from '../../dialogs/components/ApolloFormDialog'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'

const verifyContactMethodMutation = gql`
  mutation VerifyContactMethodMutation($input: VerifyContactMethodInput) {
    verifyContactMethod(input: $input) {
      contact_method_ids
    }
  }
`

const sendContactMethodVerificationMutation = gql`
  mutation SendContactMethodVerificationMutation(
    $input: SendContactMethodVerificationInput
  ) {
    sendContactMethodVerification(input: $input) {
      id
    }
  }
`

const fieldStyle = {
  width: '100%',
}

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

export default class VerificationForm extends Component {
  static propTypes = {
    id: p.string,
    open: p.bool,
    userId: p.string,
    handleRequestClose: p.func.isRequired,
  }

  constructor(props) {
    super(props)

    this.state = {
      code: '',
      submitted: false,
      readOnly: false,
      resend: false,
      sendError: '',
      loading: false,
    }
  }

  shouldSubmit = () => {
    this.setState({ submitted: true })

    const shouldSubmit = !this.getCodeError(true)
    if (shouldSubmit) {
      this.setState({ readOnly: true })
      return true
    }

    return false
  }

  sendCode = mutation => {
    this.setState({ loading: true })
    mutation({
      variables: {
        input: {
          contact_method_id: this.props.id,
        },
      },
    })
  }

  getCodeError(submitted = this.state.submitted) {
    const code = this.state.code.trim()
    if (submitted && !code) {
      return 'Code is required'
    }
    if ((submitted && code.length !== 6) || (submitted && code.match(/\D/))) {
      return 'Enter the 6-digit numeric code'
    }
  }

  getTitle() {
    if (this.state.resend) {
      return 'Resend Code'
    } else {
      return 'Send Code'
    }
  }

  renderFields() {
    const { code, loading, readOnly } = this.state

    return (
      <Grid container spacing={2}>
        <Grid item style={{ display: 'flex', alignItems: 'center' }}>
          <Mutation
            mutation={sendContactMethodVerificationMutation}
            update={() =>
              this.setState({ resend: true, sendError: '', loading: false })
            }
            onError={() =>
              this.setState({
                loading: false,
                sendError: 'Too many messages! Try again after some time.',
              })
            }
          >
            {mutation => (
              <LoadingButton
                color='primary'
                loading={loading}
                disabled={readOnly}
                buttonText={this.getTitle()}
                onClick={() => this.sendCode(mutation)}
              />
            )}
          </Mutation>
        </Grid>
        <Grid item style={{ flexGrow: 1 }}>
          <FormControl
            style={fieldStyle}
            disabled={this.state.readOnly}
            error={!!this.getCodeError()}
          >
            <TextField
              aria-label='Code'
              disabled={this.state.readOnly}
              error={!!this.getCodeError()}
              label='Verification Code'
              name='code'
              onChange={e =>
                this.setState({
                  code: e.target.value.replace(/\D/, '').slice(0, 6),
                })
              }
              placeholder='Enter the verification code received'
              value={code}
            />
            <FormHelperText>{this.getCodeError()}</FormHelperText>
          </FormControl>
        </Grid>
      </Grid>
    )
  }

  resetForm = () => {
    this.setState({
      sendError: '',
      code: '',
      submitted: false,
      readOnly: false,
      loading: false,
    })
  }

  render() {
    const { open } = this.props
    const title = 'Verify Contact Method by ' + this.props.type
    const subtitle = `Verifying "${this.props.name}" at ${formatNumber(
      this.props.value,
    )}`
    return (
      <ApolloFormDialog
        allowEdits={() => this.setState({ readOnly: false })}
        errorMessage={this.state.sendError}
        fields={this.renderFields()}
        getVariables={() => ({
          input: {
            contact_method_id: this.props.id,
            verification_code: parseInt(this.state.code),
          },
        })}
        mutation={verifyContactMethodMutation}
        onRequestClose={this.props.handleRequestClose}
        open={open}
        resetForm={this.resetForm}
        shouldSubmit={this.shouldSubmit}
        subtitle={subtitle}
        title={title}
      />
    )
  }
}
