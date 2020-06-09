import React from 'react'
import { PropTypes as p } from 'prop-types'
import { FormContainer, FormField } from '../forms'
import { UserSelect } from '../selection'
import { Grid } from '@material-ui/core'

export default function UserForm(props) {
  return (
    <FormContainer {...props}>
      <Grid container spacing={2}>
        <FormField
          component={UserSelect}
          disabled={false}
          fieldName='users'
          fullWidth
          label='Select User(s)'
          multiple
          name='users'
          required
          value={props.value.users}
        />
      </Grid>
    </FormContainer>
  )
}

UserForm.PropTypes = {
  errors: p.array,
  onChange: p.func,
  disabled: p.bool,
  value: p.shape({
    users: p.arrayOf(p.string),
  }).isRequired,
}
