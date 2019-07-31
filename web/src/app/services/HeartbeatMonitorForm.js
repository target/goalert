import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import withStyles from '@material-ui/core/styles/withStyles'
import { FormContainer, FormField } from '../forms'

const styles = theme => ({
  infoIcon: {
    color: theme.palette.primary['500'],
  },
})

const clampTimeout = val => {
  const num = parseInt(val, 10)
  if (Number.isNaN(num)) return val

  // need to have the min be 1 here so you can type `10`
  return Math.min(Math.max(1, num), 9000)
}

@withStyles(styles)
export default class HeartbeatMonitorForm extends React.PureComponent {
  static propTypes = {
    value: p.shape({
      name: p.string.isRequired,
      timeoutMinutes: p.number.isRequired,
    }).isRequired,

    errors: p.arrayOf(
      p.shape({
        field: p.oneOf(['name', 'timeoutMinutes']).isRequired,
        message: p.string.isRequired,
      }),
    ),

    onChange: p.func,
  }

  render() {
    const { classes, ...formProps } = this.props
    return (
      <FormContainer {...formProps} optionalLabels>
        <Grid container spacing={2}>
          <Grid item style={{ flexGrow: 1 }} xs={12}>
            <FormField
              fullWidth
              component={TextField}
              label='Name'
              name='name'
              required
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              required
              type='number'
              label='Timeout (minutes)'
              name='timeoutMinutes'
              min={5}
              max={9000}
              mapOnChangeValue={clampTimeout}
            />
          </Grid>
        </Grid>
      </FormContainer>
    )
  }
}
