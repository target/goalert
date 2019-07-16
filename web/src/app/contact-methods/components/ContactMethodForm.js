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
import { withApollo } from 'react-apollo'
import ApolloFormDialog from '../../dialogs/components/ApolloFormDialog'
import { createNotificationRuleMutation } from '../../notification-rules/components/CreateNotificationRuleForm'
import UserContactMethodVerificationDialog from '../../users/UserContactMethodVerificationDialog'

const createContactMethodMutation = gql`
  mutation CreateContactMethodMutation($input: CreateContactMethodInput) {
    createContactMethod(input: $input) {
      id
      name
      type
      value
      disabled
    }
  }
`

const updateContactmethodMutation = gql`
  mutation UpdateContactMethodMutation($input: UpdateContactMethodInput) {
    updateContactMethod(input: $input) {
      id
      name
      type
      value
      disabled
    }
  }
`

const types = ['SMS', 'VOICE']
const countryCodeOptions = [
  {
    label: '+1 (United States of America)',
    value: '+1',
    length: 10,
  },
  {
    label: '+91 (India)',
    value: '+91',
    length: 10,
  },
  {
    label: '+44 (United Kingdom)',
    value: '+44',
    length: 10,
  },
]

const fieldStyle = {
  width: '100%',
}

const getPhoneLen = code =>
  countryCodeOptions.find(o => o.value === code).length

const splitNumber = phone => {
  const cc = countryCodeOptions.find(o => phone.startsWith(o.value))
  if (!cc) {
    throw new Error('invalid or unknown country code for number: ' + phone)
  }
  return {
    cc: cc.value,
    phone: phone.slice(cc.value.length),
  }
}

class ContactMethodForm extends Component {
  static propTypes = {
    id: p.string,
    type: p.string,
    value: p.string,
    name: p.string,
    open: p.bool,
    userId: p.string,
    existing: p.arrayOf(
      p.shape({
        id: p.string.isRequired,
        type: p.string.isRequired,
        name: p.string.isRequired,
      }),
    ).isRequired,
    handleRequestClose: p.func.isRequired,
    cmDisabled: p.bool,
  }

  constructor(props) {
    super(props)

    const type = props.type || 'SMS'
    const value = props.value || ''

    let phone = ''
    let cc = '+1'

    if ((type === 'SMS' || type === 'VOICE') && value.length > 2) {
      const n = splitNumber(value)
      phone = n.phone
      cc = n.cc
    }

    this.state = {
      name: props.name || '',
      type: type,
      countryCode: cc,
      phone: phone,
      submitted: false,
      readOnly: false,
      contactMethod: {},
    }
  }

  shouldSubmit = () => {
    this.setState({ submitted: true })

    const shouldSubmit = !(this.getNameError(true) || this.getValueError())
    if (shouldSubmit) {
      this.setState({ readOnly: true })
      return true
    }

    return false
  }

  getVariables = () => {
    if (this.props.id) {
      return {
        input: {
          id: this.props.id,
          disabled: this.props.cmDisabled,
          name: this.state.name,
          type: this.state.type,
          value: this.getValue(),
        },
      }
    } else {
      return {
        input: {
          user_id: this.props.userId,
          name: this.state.name,
          type: this.state.type,
          value: this.getValue(),
        },
      }
    }
  }

  // update cache
  createNotificationRule = cm => {
    return this.props.client
      .mutate({
        mutation: createNotificationRuleMutation,
        variables: {
          input: {
            user_id: this.props.userId,
            delay_minutes: 0,
            contact_method_id: cm.id,
          },
        },
      })
      .catch(err => console.error(err))
  }

  onCreateCMSuccess = (cache, data) => {
    this.setState({ submitted: false })
    const cm = data.createContactMethod
    if (!cm) return // don't need to update cache on an update vs. create
    this.setState({ contactMethod: cm, showVerifyForm: cm.disabled })

    // if contact method enabled and no notification rules, create notification rule
    if (
      !cm.disabled &&
      (!this.props.notificationRules ||
        this.props.notificationRules.length === 0)
    ) {
      this.createNotificationRule(cm)
    }
  }

  onVerificationSuccess = () => {
    // create notification that notifies immediately for new user's contact method
    this.createNotificationRule(this.state.contactMethod)
  }

  renderVerificationForm = () =>
    this.state.showVerifyForm && (
      <UserContactMethodVerificationDialog
        onClose={() => this.setState({ showVerifyForm: false })}
        contactMethodID={this.state.contactMethod.id}
      />
    )

  getValue() {
    switch (this.state.type) {
      case 'SMS':
      case 'VOICE':
        return this.state.countryCode + this.state.phone
    }
  }

  getNameError(submitted = this.state.submitted) {
    const name = this.state.name.trim()
    if (
      submitted &&
      this.props.existing.some(
        e =>
          e.type === this.state.type &&
          e.id !== this.props.id &&
          e.name === name,
      )
    ) {
      return 'Name must be unique for a given type.'
    }

    if (submitted && !name) {
      return 'A name is required.'
    }
  }

  getValueError() {
    switch (this.state.type) {
      case 'SMS':
      case 'VOICE':
        return this.getPhoneError(true)
    }
  }

  getPhoneError(submitted = this.state.submitted) {
    // The only invalid case is too few digits, since we filter inputs.
    const len = getPhoneLen(this.state.countryCode)
    if (submitted && this.state.phone.length < len) {
      return 'Enter a ' + len + ' digit number (including area code).'
    }
  }

  filterSetPhone(newVal) {
    this.setState({ phone: newVal.replace(/[^0-9]/g, '').slice(0, 10) })
  }

  renderFields() {
    const { name, type, countryCode } = this.state

    let selectField = (
      <Select
        name='type'
        value={type}
        aria-label='Method Type'
        renderValue={value => `${value}`}
        disabled={!!this.props.id}
        onChange={event => this.setState({ type: event.target.value })}
      >
        {types.map(type => {
          return (
            <MenuItem key={type} value={type}>
              {type}
            </MenuItem>
          )
        })}
      </Select>
    )

    return (
      <Grid container spacing={2}>
        <Grid item xs={12} sm={6}>
          <FormControl style={fieldStyle} error={!!this.getNameError()}>
            <TextField
              aria-label='Name'
              disabled={this.state.readOnly}
              error={!!this.getNameError()}
              label='Contact Method Name'
              name='name'
              onChange={e => this.setState({ name: e.target.value })}
              placeholder='Personal, Work, Home...'
              value={name}
            />
            <FormHelperText>{this.getNameError()}</FormHelperText>
          </FormControl>
        </Grid>
        <Grid item xs={12} sm={6}>
          <FormControl
            style={fieldStyle}
            disabled={this.state.readOnly}
            label='Type'
          >
            <InputLabel htmlFor='Type'>Type</InputLabel>
            {selectField}
          </FormControl>
        </Grid>
        <Grid item xs={12} sm={6}>
          <FormControl disabled={this.state.readOnly} style={fieldStyle}>
            <InputLabel htmlFor='Country Code'>Country Code</InputLabel>
            <Select
              name='countryCode'
              aria-label='Country Code'
              value={countryCode}
              renderValue={value =>
                `${countryCodeOptions.find(option => option.value === value)
                  .label || value}`
              }
              onChange={event =>
                this.setState({ countryCode: event.target.value })
              }
            >
              {countryCodeOptions.map(option => {
                return (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label || option.value}
                  </MenuItem>
                )
              })}
            </Select>
          </FormControl>
        </Grid>
        <Grid item xs={12} sm={6}>
          <FormControl style={fieldStyle} error={!!this.getPhoneError()}>
            <TextField
              aria-label='Phone Number'
              disabled={this.state.readOnly}
              error={!!this.getPhoneError()}
              label='Phone Number'
              name='phone'
              onChange={e => this.filterSetPhone(e.target.value)}
              type='tel'
              value={this.state.phone}
            />
            <FormHelperText>{this.getPhoneError()}</FormHelperText>
          </FormControl>
        </Grid>
      </Grid>
    )
  }

  resetForm = () => {
    const type = this.props.type || 'SMS'
    const value = this.props.value || ''

    let phone = ''
    let cc = '+1'

    if ((type === 'SMS' || type === 'VOICE') && value.length > 2) {
      const n = splitNumber(value)
      phone = n.phone
      cc = n.cc
    }

    this.setState({
      name: this.props.name || '',
      type: type,
      countryCode: cc,
      phone: phone,
      submitted: false,
      readOnly: false,
    })
  }

  render() {
    const { open, newUser, id } = this.props

    const newUserText = 'To get started, please enter a contact method.'
    const newUserCaption =
      'By entering your contact information, you agree to receive auto-dialed ' +
      'and prerecorded alert calls or texts from Target or those acting on behalf of Target Corporation.'

    let title = 'Add New Contact Method'
    if (newUser) {
      title = 'Welcome to GoAlert!'
    } else if (id) {
      title = 'Edit Contact Method'
    }

    return (
      <React.Fragment>
        <ApolloFormDialog
          allowEdits={() => this.setState({ readOnly: false })}
          caption={newUser ? newUserCaption : null}
          disableCancel={newUser}
          fields={this.renderFields()}
          getVariables={this.getVariables}
          mutation={
            id ? updateContactmethodMutation : createContactMethodMutation
          }
          onRequestClose={this.props.handleRequestClose}
          onSuccess={this.onCreateCMSuccess}
          open={open}
          resetForm={this.resetForm}
          shouldSubmit={this.shouldSubmit}
          subtitle={newUser ? newUserText : null}
          title={title}
        />
        {this.renderVerificationForm()}
      </React.Fragment>
    )
  }
}

export default withApollo(ContactMethodForm)
