import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import FormControl from '@material-ui/core/FormControl'
import FormHelperText from '@material-ui/core/FormHelperText'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import gql from 'graphql-tag'
import { ServiceSelect } from '../../selection'
import ApolloFormDialog from '../../dialogs/components/ApolloFormDialog'

const mutation = gql`
  mutation CreateAlertMutation($input: CreateAlertInput!) {
    createAlert(input: $input) {
      number: _id
      id
      status: status_2
      created_at
      escalation_level
      description
      details
      summary
      service_id
      source
      assignments {
        id
        name
      }
      service {
        id
        name
        escalation_policy_id
      }
      logs_2 {
        event
        message
        timestamp
      }
      escalation_policy_snapshot {
        repeat
        current_level
        last_escalation
        steps {
          delay_minutes
          users {
            id
            name
          }
          schedules {
            id
            name
          }
        }
      }
    }
  }
`

export default class AlertForm extends Component {
  static contextTypes = {
    router: p.object,
  }

  constructor(props) {
    super(props)

    let sid
    if (props.service) sid = props.service.id

    this.state = {
      sid,
      errorMessage: '',
      submitted: false,
      summary: '',
      details: '',
      readOnly: false,
    }
  }

  shouldSubmit = () => {
    this.setState({ submitted: true })

    const shouldSubmit = !this.validateForm()
    if (shouldSubmit) {
      this.setState({ readOnly: true })
      return true
    }

    return false
  }

  onSuccess = (cache, data) => {
    const alert = data.createAlert

    // Get created alert ID back from the promise from mutation
    const alertPage = '/alerts/' + encodeURIComponent(alert.number)

    // Redirect to created alert's page
    this.context.router.history.push(alertPage)
  }

  getVariables = () => {
    return {
      input: {
        service_id: this.state.sid,
        description:
          this.state.summary.trim() + '\n' + this.state.details.trim(),
      },
    }
  }

  validateForm() {
    return this.validateService(true) || this.validateSummary(true)
  }

  validateService(submitted = this.state.submitted) {
    if (!submitted) return ''
    if (!this.state.sid) return 'A service must be selected'
    return ''
  }

  validateSummary(submitted = this.state.submitted) {
    if (!submitted) return ''
    if (this.state.summary.length < 2) {
      return 'Summary must be at least 2 characters'
    }
    return ''
  }

  renderServiceField() {
    if (this.props.service) {
      return (
        <TextField
          style={{ width: '100%' }}
          aria-label='Select Service'
          label='Select Service'
          disabled
          value={this.props.service.name}
        />
      )
    }

    return (
      <ServiceSelect
        value={this.state.sid || ''}
        onChange={value => this.setState({ sid: value })}
        label='Select Service'
        name='service'
        errorMessage={this.validateService()}
        disabled={this.state.readOnly}
      />
    )
  }

  resetForm = () => {
    const { service } = this.props

    // Reset the form when we open it.
    this.setState({
      sid: service ? service.id : null,
      errorMessage: '',
      submitted: false,
      summary: '',
      details: '',
      readOnly: false,
    })
  }

  render() {
    const { open } = this.props

    const formFields = (
      <Grid container spacing={16}>
        <Grid item xs={12}>
          <FormControl
            error={!!this.validateSummary()}
            style={{ width: '100%' }}
          >
            <TextField
              aria-label='Alert Summary'
              disabled={this.state.readOnly}
              error={!!this.validateSummary()}
              label='Alert Summary'
              name='summary'
              onChange={event => this.setState({ summary: event.target.value })}
            />
            <FormHelperText>{this.validateSummary()}</FormHelperText>
          </FormControl>
        </Grid>
        <Grid item xs={12}>
          <FormControl style={{ width: '100%' }}>
            <TextField
              name='details'
              aria-label='Alert Details'
              label='Alert Details'
              multiline
              rowsMax={10}
              disabled={this.state.readOnly}
              onChange={event => this.setState({ details: event.target.value })}
            />
          </FormControl>
        </Grid>
        <Grid item xs={12}>
          {this.renderServiceField()}
        </Grid>
      </Grid>
    )

    return (
      <ApolloFormDialog
        allowEdits={() => this.setState({ readOnly: false })}
        fields={formFields}
        getVariables={this.getVariables}
        mutation={mutation}
        onRequestClose={this.props.handleRequestClose}
        resetForm={this.resetForm}
        open={open}
        shouldSubmit={this.shouldSubmit}
        onSuccess={this.onSuccess}
        title='Create New Alert'
      />
    )
  }
}
