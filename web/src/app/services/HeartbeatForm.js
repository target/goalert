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

@withStyles(styles)
export default class HeartbeatForm extends React.PureComponent {
  static propTypes = {
    value: p.shape({
      name: p.string,
      duration: p.number,
    }).isRequired,

    errors: p.arrayOf(
      p.shape({
        field: p.oneOf(['name', 'duration']).isRequired,
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
              label='Duration'
              name='duration'
            />
          </Grid>
        </Grid>
      </FormContainer>
    )
  }
}
