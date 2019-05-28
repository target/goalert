import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { FormContainer, FormField } from '../forms'
import UserContactMethodSelect from './UserContactMethodSelect'

export default class UserNotificationRuleForm extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,

    value: p.shape({
      contactMethodID: p.string.isRequired,
      delayMinutes: p.number.isRequired,
    }).isRequired,

    errors: p.arrayOf(
      p.shape({
        field: p.oneOf(['delayMinutes', 'contactMethodID']).isRequired,
        message: p.string.isRequired,
      }),
    ),

    onChange: p.func,

    disabled: p.bool,
  }

  render() {
    const { userID, ...other } = this.props
    return (
      <FormContainer {...other} optionalLabels>
        <Grid container spacing={16}>
          <Grid item xs={12} sm={12} md={6}>
            <FormField
              fullWidth
              name='contactMethodID'
              label='Contact Method'
              userID={userID}
              required
              component={UserContactMethodSelect}
            />
          </Grid>
          <Grid item xs={12} sm={12} md={6}>
            <FormField
              fullWidth
              name='delayMinutes'
              required
              label='Delay (minutes)'
              type='number'
              component={TextField}
            />
          </Grid>
        </Grid>
      </FormContainer>
    )
  }
}
