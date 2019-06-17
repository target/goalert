import React, { Component } from 'react'
import p from 'prop-types'
import FormControl from '@material-ui/core/FormControl'
import FormHelperText from '@material-ui/core/FormHelperText'
import Grid from '@material-ui/core/Grid'
import InputLabel from '@material-ui/core/InputLabel'
import MenuItem from '@material-ui/core/MenuItem'
import Select from '@material-ui/core/Select'
import TextField from '@material-ui/core/TextField'
import gql from 'graphql-tag'
import ApolloFormDialog from '../../dialogs/components/ApolloFormDialog'

const nrPrefix = delay =>
  delay ? `If I do not respond after ${delay} minute(s)` : 'Immediately'
const cmText = cm => {
  switch (cm.type) {
    case 'SMS':
      return `send an SMS to my ${cm.name} number`
    case 'VOICE':
      return `call my ${cm.name} number`
  }
}

export const createNotificationRuleMutation = gql`
  mutation CreateNotificationRuleMutation($input: CreateNotificationRuleInput) {
    createNotificationRule(input: $input) {
      id
      delay
      delay_minutes
      contact_method_id
      contact_method {
        id
        name
        type
        value
        disabled
      }
    }
  }
`

class CreateNotificationRuleForm extends Component {
  static propTypes = {
    userId: p.string.isRequired,
  }

  constructor(props) {
    super(props)

    let cm
    let delay
    for (delay = 0; delay < 50; delay += 5) {
      cm = props.contactMethods.find(cm => !this.cmError(false, delay, cm.id))
      if (cm) {
        break
      }
    }

    if (!cm) cm = props.contactMethods[0]

    this.state = {
      delay: delay || '',
      cm: cm.id,
      errorMessage: '',
      submitted: false,
      readOnly: false,
    }
  }

  getVariables = () => {
    return {
      input: {
        user_id: this.props.userId,
        delay_minutes: this.state.delay || 0,
        contact_method_id: this.state.cm,
      },
    }
  }

  shouldSubmit = () => {
    this.setState({ submitted: true })

    const shouldSubmit = !this.cmError(true)
    if (shouldSubmit) {
      this.setState({ readOnly: true })
      return true
    }

    return false
  }

  cmError(
    submitted = this.state.submitted,
    delay = this.state.delay,
    cm = this.state.cm,
  ) {
    // check that a rule doesn't exist for this cm and delay
    if (
      submitted &&
      this.props.rules.some(
        r => r.delay === delay && r.contact_method_id === cm,
      )
    ) {
      return 'Contact method is already used for the given delay.'
    }
  }

  setDelay(v) {
    const n = parseInt(v, 10)

    // Allow the user to backspace all the characters
    if (Number.isNaN(n)) {
      this.setState({ delay: '' })
      return
    }

    if (n < 0) {
      this.setState({ delay: 0 })
      return
    }

    if (n > 9000) {
      this.setState({ delay: 9000 })
      return
    }

    this.setState({ delay: n })
  }

  contactOptions() {
    return this.props.contactMethods.map(cm => {
      let name = cmText(cm)
      name = name.charAt(0).toUpperCase() + name.slice(1)

      return {
        id: cm.id,
        label: name,
        value: name,
      }
    })
  }

  renderFields = () => {
    const opts = this.contactOptions()
    return (
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormControl error={!!this.cmError()} style={{ width: '100%' }}>
            <TextField
              aria-label='Delay Time'
              disabled={this.state.readOnly}
              error={!!this.cmError()}
              label='Delay Time'
              name='delay'
              min='0'
              onChange={e => this.setDelay(e.target.value)}
              type='number'
              value={this.state.delay}
            />
            <FormHelperText>
              Notify: {nrPrefix(this.state.delay)}
            </FormHelperText>
          </FormControl>
        </Grid>
        <Grid item xs={12}>
          <FormControl
            disabled={this.state.readOnly}
            error={!!this.cmError()}
            style={{ width: '100%' }}
          >
            <InputLabel htmlFor='Action'>Action</InputLabel>
            <Select
              name='contact'
              aria-label='Then'
              label='Then'
              value={this.state.cm}
              renderValue={value => `${opts.find(o => o.id === value).label}`}
              onChange={event => this.setState({ cm: event.target.value })}
            >
              {opts.map(option => {
                return (
                  <MenuItem key={option.id} value={option.id}>
                    {option.label}
                  </MenuItem>
                )
              })}
            </Select>
            <FormHelperText>{this.cmError()}</FormHelperText>
          </FormControl>
        </Grid>
      </Grid>
    )
  }

  resetForm = () => {
    // Reset the form when we open it.
    let cm
    let delay
    for (delay = 0; delay < 50; delay += 5) {
      cm = this.props.contactMethods.find(
        cm => !this.cmError(false, delay, cm.id),
      )
      if (cm) {
        break
      }
    }

    if (!cm) cm = this.props.contactMethods[0]

    this.setState({
      delay: delay || '',
      cm: cm.id,
      errorMessage: '',
      submitted: false,
      readOnly: false,
    })
  }

  render() {
    const { open } = this.props

    return (
      <ApolloFormDialog
        allowEdits={() => this.setState({ readOnly: false })}
        fields={this.renderFields()}
        getVariables={this.getVariables}
        mutation={createNotificationRuleMutation}
        onRequestClose={this.props.handleRequestClose}
        open={open}
        resetForm={this.resetForm}
        shouldSubmit={this.shouldSubmit}
        title='Add New Notification Rule'
      />
    )
  }
}

export default CreateNotificationRuleForm
