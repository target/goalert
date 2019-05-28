import React, { PureComponent } from 'react'
import { PropTypes as p } from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { FormContainer, FormField } from '../forms'
import MaterialSelect from '../selection/MaterialSelect'

export default class PolicyForm extends PureComponent {
  static propTypes = {
    value: p.shape({
      name: p.string,
      description: p.string,
      repeat: p.shape({
        label: p.string.isRequired,
        value: p.string.isRequired,
      }).isRequired,
    }).isRequired,

    errors: p.arrayOf(
      p.shape({
        field: p.oneOf(['name', 'description', 'repeat']).isRequired,
        message: p.string.isRequired,
      }),
    ),

    disabled: p.bool,
    onChange: p.func,
  }

  render() {
    return (
      <FormContainer optionalLabels {...this.props}>
        <Grid container spacing={16}>
          <Grid item xs={12}>
            <FormField
              component={TextField}
              disabled={this.props.disabled}
              fieldName='name'
              fullWidth
              label='Name'
              name='name'
              required
              value={this.props.value.name}
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              component={TextField}
              disabled={this.props.disabled}
              fieldName='description'
              fullWidth
              label='Description'
              multiline
              name='description'
              value={this.props.value.description}
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              component={MaterialSelect}
              disabled={this.props.disabled}
              fieldName='repeat'
              fullWidth
              hint='The amount of times it will escalate through all steps'
              label='Repeat Count'
              name='repeat'
              options={[
                { label: '0', value: '0' },
                { label: '1', value: '1' },
                { label: '2', value: '2' },
                { label: '3', value: '3' },
                { label: '4', value: '4' },
                { label: '5', value: '5' },
              ]}
              required
              value={this.props.value.repeat.value}
            />
          </Grid>
        </Grid>
      </FormContainer>
    )
  }
}
