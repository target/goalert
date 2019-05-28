import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { EscalationPolicySelect } from '../selection/EscalationPolicySelect'
import { FormContainer, FormField } from '../forms'

export default class ServiceForm extends React.PureComponent {
  static propTypes = {
    value: p.shape({
      name: p.string,
      description: p.string,
      escalationPolicyID: p.string,
    }).isRequired,

    // indicates that the escalation policy is a required field
    epRequired: p.bool,

    errors: p.arrayOf(
      p.shape({
        field: p.oneOf(['name', 'description', 'escalationPolicyID'])
          .isRequired,
        message: p.string.isRequired,
      }),
    ),

    onChange: p.func,

    disabled: p.bool,
  }

  fieldError(name) {
    if (!this.props.errors) return null
    const err = this.props.errors.find(err => err.field === name)
    return err ? err.message : null
  }

  onChange = values => {
    if (!this.props.onChange) return
    this.props.onChange({
      ...this.props.value,
      ...values,
    })
  }

  render() {
    const { epRequired, ...containerProps } = this.props

    return (
      <FormContainer {...containerProps} optionalLabels={this.props.epRequired}>
        <Grid container spacing={16}>
          <Grid item xs={12}>
            <FormField
              fullWidth
              label='Name'
              name='name'
              required
              component={TextField}
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              label='Description'
              name='description'
              multiline
              component={TextField}
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              label='Escalation Policy'
              name='escalation-policy'
              fieldName='escalationPolicyID'
              required={epRequired}
              component={EscalationPolicySelect}
            />
          </Grid>
        </Grid>
      </FormContainer>
    )
  }
}
