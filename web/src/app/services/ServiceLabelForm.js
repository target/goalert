import React from 'react'
import p from 'prop-types'
import { FormContainer, FormField } from '../forms'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { LabelKeySelect } from '../selection/LabelKeySelect'
import { Config } from '../util/RequireConfig'

export default class LabelForm extends React.PureComponent {
  static propTypes = {
    value: p.shape({
      key: p.string.isRequired,
      value: p.string.isRequired,
    }).isRequired,

    errors: p.arrayOf(
      p.shape({
        field: p.oneOf(['key', 'value']).isRequired,
        message: p.string.isRequired,
      }),
    ),

    onChange: p.func,

    editValueOnly: p.bool,
    create: p.bool,
  }

  static defaultProps = {
    editValueOnly: false,
  }

  validateKey = value => {
    const parts = value.split('/')
    if (parts.length !== 2)
      return new Error('Must be in the format: "example/KeyName".')
  }

  render() {
    const { editValueOnly, create, ...otherProps } = this.props

    return (
      <FormContainer {...otherProps} optionalLabels>
        <Grid container spacing={2}>
          <Grid item xs={12}>
            <Config>
              {cfg => (
                <FormField
                  fullWidth
                  disabled={editValueOnly}
                  component={LabelKeySelect}
                  label='Key'
                  name='key'
                  required
                  onCreate={
                    // if create is enabled, allow new keys to be provided
                    !cfg['General.DisableLabelCreation'] &&
                    (key => otherProps.onChange({ ...otherProps.value, key }))
                  }
                  validate={this.validateKey}
                />
              )}
            </Config>
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              label='Value'
              name='value'
              required
            />
          </Grid>
        </Grid>
      </FormContainer>
    )
  }
}
